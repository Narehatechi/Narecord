/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var ReleaseData GithubRelease
var GithubError error
var GithubDoneChan chan bool

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

// Equicord publishes desktop.asar on the latest release, so one latest-endpoint
// hit is the common path. Walk at most two list pages if it is missing. Cache the
// resolved release for process lifetime so Install / Repair does not re-hit GitHub.
const desktopAsarListMaxPages = 2

var (
	installedHashRe         = regexp.MustCompile(`// (?:Narecord|Equicord|Vencord) (\w+)`)
	desktopAsarCacheMu      sync.Mutex
	desktopAsarReleaseCache *GithubRelease
)

func resetDesktopAsarReleaseCache() {
	desktopAsarCacheMu.Lock()
	desktopAsarReleaseCache = nil
	desktopAsarCacheMu.Unlock()
}

func desktopAsarDownloadURL(rel *GithubRelease) string {
	if rel == nil {
		return ""
	}
	for _, ass := range rel.Assets {
		if ass.Name == "desktop.asar" {
			return ass.DownloadURL
		}
	}
	return ""
}

func pickReleaseWithDesktopAsar(latest *GithubRelease, rest []GithubRelease) *GithubRelease {
	if desktopAsarDownloadURL(latest) != "" {
		return latest
	}
	for i := range rest {
		if desktopAsarDownloadURL(&rest[i]) != "" {
			return &rest[i]
		}
	}
	return nil
}

func releaseHash(rel *GithubRelease) string {
	if rel == nil {
		return "Unknown"
	}
	if i := strings.LastIndex(rel.Name, " "); i >= 0 && i+1 < len(rel.Name) {
		return rel.Name[i+1:]
	}
	if rel.TagName != "" {
		return rel.TagName
	}
	if rel.Name != "" {
		return rel.Name
	}
	return "Unknown"
}

func githubGet(url, fallbackUrl string) (*http.Response, error) {
	Log.Debug("Fetching", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error("Failed to create Request", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Log.Error("Failed to send Request", err)
		return nil, err
	}

	if res.StatusCode >= 300 {
		isRateLimitedOrBlocked := res.StatusCode == 401 || res.StatusCode == 403 || res.StatusCode == 429
		triedFallback := url == fallbackUrl || fallbackUrl == ""
		_ = res.Body.Close()

		if isRateLimitedOrBlocked && !triedFallback {
			Log.Error(fmt.Sprintf("Failed to fetch %s (status code %d). Trying fallback url %s", url, res.StatusCode, fallbackUrl))
			return githubGet(fallbackUrl, fallbackUrl)
		}

		err = errors.New(res.Status)
		Log.Error(url, "returned Non-OK status", err)
		return nil, err
	}

	return res, nil
}

func GetGithubRelease(url, fallbackUrl string) (*GithubRelease, error) {
	res, err := githubGet(url, fallbackUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data GithubRelease
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub JSON Response", err)
		return nil, err
	}

	return &data, nil
}

func GetGithubReleases(url, fallbackUrl string) ([]GithubRelease, error) {
	res, err := githubGet(url, fallbackUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data []GithubRelease
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub releases JSON Response", err)
		return nil, err
	}

	return data, nil
}

func releasesPageURL(base string, page int) string {
	return fmt.Sprintf("%s?per_page=20&page=%d", base, page)
}

func fetchReleaseWithDesktopAsar() (*GithubRelease, error) {
	desktopAsarCacheMu.Lock()
	defer desktopAsarCacheMu.Unlock()
	if desktopAsarDownloadURL(desktopAsarReleaseCache) != "" {
		return desktopAsarReleaseCache, nil
	}

	rel, err := fetchReleaseWithDesktopAsarUncached()
	if err != nil {
		return nil, err
	}
	desktopAsarReleaseCache = rel
	return rel, nil
}

func fetchReleaseWithDesktopAsarUncached() (*GithubRelease, error) {
	latest, latestErr := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
	if latestErr == nil && desktopAsarDownloadURL(latest) != "" {
		return latest, nil
	}
	if latestErr != nil {
		Log.Debug("Latest release fetch failed:", latestErr)
		// Rate-limited / blocked: do not walk extra list pages.
		if isGithubRateLimitStatus(latestErr) {
			return nil, latestErr
		}
	} else {
		Log.Debug("Latest release", latest.TagName, "has no desktop.asar; walking recent releases")
	}

	var listErr error
	for page := 1; page <= desktopAsarListMaxPages; page++ {
		list, err := GetGithubReleases(releasesPageURL(ReleaseListUrl, page), releasesPageURL(ReleaseListUrlFallback, page))
		if err != nil {
			listErr = err
			break
		}
		if picked := pickReleaseWithDesktopAsar(nil, list); picked != nil {
			Log.Debug("Found desktop.asar on release", picked.TagName)
			return picked, nil
		}
		if len(list) == 0 {
			break
		}
	}
	if isGithubRateLimitStatus(listErr) {
		return nil, listErr
	}

	fallback, fbErr := GetGithubRelease(DesktopAsarFallbackUrl, DesktopAsarFallbackUrl)
	if fbErr == nil && desktopAsarDownloadURL(fallback) != "" {
		Log.Debug("Using known desktop.asar fallback", fallback.TagName)
		return fallback, nil
	}

	if latestErr != nil {
		return nil, latestErr
	}
	if listErr != nil {
		return nil, listErr
	}
	if fbErr != nil {
		return nil, fbErr
	}
	return nil, errors.New("Didn't find desktop.asar download link")
}

func isGithubRateLimitStatus(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "401") || strings.Contains(s, "403") || strings.Contains(s, "429")
}

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("NARECORD_DEV_INSTALL") == "1" || os.Getenv("EQUICORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	go func() {
		// Make sure UI updates once the request either finished or failed
		defer func() {
			GithubDoneChan <- GithubError == nil
		}()

		data, err := fetchReleaseWithDesktopAsar()
		if err != nil {
			GithubError = err
			return
		}

		ReleaseData = *data
		LatestHash = releaseHash(data)
		Log.Debug("Finished fetching GitHub Data")
		Log.Debug("Latest hash is", LatestHash, "Local Install is", Ternary(LatestHash == InstalledHash, "up to date!", "outdated!"))
	}()

	// either .asar file or directory with main.js file (in DEV)
	NarecordFile := NarecordDirectory

	stat, err := os.Stat(NarecordFile)
	if err != nil {
		return
	}

	// dev
	if stat.IsDir() {
		NarecordFile = path.Join(NarecordFile, "main.js")
	}

	// Check hash of installed version if exists
	b, err := os.ReadFile(NarecordFile)
	if err != nil {
		return
	}

	Log.Debug("Found existing Narecord Install. Checking for hash...")

	match := installedHashRe.FindSubmatch(b)
	if match != nil {
		InstalledHash = string(match[1])
		Log.Debug("Existing hash is", InstalledHash)

	} else {
		Log.Debug("Didn't find hash")

	}
}

func installLatestBuilds() (retErr error) {
	Log.Debug("Installing latest builds...")

	if IsDevInstall {
		Log.Debug("Skipping due to dev install")
		return
	}

	downloadUrl := desktopAsarDownloadURL(&ReleaseData)
	if downloadUrl == "" {
		data, err := fetchReleaseWithDesktopAsar()
		if err != nil {
			retErr = errors.New("Didn't find desktop.asar download link")
			Log.Error(retErr)
			return
		}
		ReleaseData = *data
		LatestHash = releaseHash(data)
		downloadUrl = desktopAsarDownloadURL(&ReleaseData)
	}

	if downloadUrl == "" {
		retErr = errors.New("Didn't find desktop.asar download link")
		Log.Error(retErr)
		return
	}

	Log.Debug("Downloading desktop.asar")

	res, err := http.Get(downloadUrl)
	if err == nil && res.StatusCode >= 300 {
		err = errors.New(res.Status)
	}
	if err != nil {
		if res != nil {
			_ = res.Body.Close()
		}
		Log.Error("Failed to download desktop.asar:", err)
		retErr = err
		return
	}
	defer res.Body.Close()
	out, err := os.OpenFile(NarecordDirectory, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		Log.Error("Failed to create", NarecordDirectory+":", err)
		retErr = err
		return
	}
	defer out.Close()
	read, err := io.Copy(out, res.Body)
	if err != nil {
		Log.Error("Failed to download to", NarecordDirectory+":", err)
		retErr = err
		return
	}
	contentLength := res.Header.Get("Content-Length")
	expected := strconv.FormatInt(read, 10)
	if expected != contentLength {
		err = errors.New("Unexpected end of input. Content-Length was " + contentLength + ", but I only read " + expected)
		Log.Error(err.Error())
		retErr = err
		return
	}

	_ = FixOwnership(NarecordDirectory)

	InstalledHash = LatestHash
	return
}
