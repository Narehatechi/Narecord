/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestPickReleaseWithDesktopAsarSkipsInstallerOnlyLatest(t *testing.T) {
	latest := &GithubRelease{
		Name:    "Narecord Installer v1.1.0",
		TagName: "v1.1.0",
		Assets: []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{
			{Name: "NarecordInstaller.exe", DownloadURL: "https://example.test/NarecordInstaller.exe"},
		},
	}
	older := GithubRelease{
		Name:    "Narecord Installer v1.0.0",
		TagName: "v1.0.0",
		Assets: []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{
			{Name: "desktop.asar", DownloadURL: "https://example.test/desktop.asar"},
		},
	}

	if got := desktopAsarDownloadURL(latest); got != "" {
		t.Fatalf("latest should not have desktop.asar, got %s", got)
	}

	picked := pickReleaseWithDesktopAsar(latest, []GithubRelease{older})
	if picked == nil {
		t.Fatal("expected to find desktop.asar on an older release")
	}
	if picked.TagName != "v1.0.0" {
		t.Fatalf("picked %s, want v1.0.0", picked.TagName)
	}
	if got := desktopAsarDownloadURL(picked); got != "https://example.test/desktop.asar" {
		t.Fatalf("download url = %s", got)
	}
}

func TestPickReleaseWithDesktopAsarPrefersLatestWhenPresent(t *testing.T) {
	latest := &GithubRelease{
		TagName: "v2.0.0",
		Assets: []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{
			{Name: "desktop.asar", DownloadURL: "https://example.test/new.asar"},
		},
	}
	older := GithubRelease{
		TagName: "v1.0.0",
		Assets: []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{
			{Name: "desktop.asar", DownloadURL: "https://example.test/old.asar"},
		},
	}
	picked := pickReleaseWithDesktopAsar(latest, []GithubRelease{older})
	if picked == nil || picked.TagName != "v2.0.0" {
		t.Fatalf("picked %+v, want latest v2.0.0", picked)
	}
}

func TestReleaseHash(t *testing.T) {
	if got := releaseHash(&GithubRelease{Name: "Narecord abcdef1"}); got != "abcdef1" {
		t.Fatalf("got %q", got)
	}
	if got := releaseHash(&GithubRelease{TagName: "v1.0.0"}); got != "v1.0.0" {
		t.Fatalf("got %q", got)
	}
}

func TestDesktopAsarURLsPreferNarecordWithEquicordFallback(t *testing.T) {
	if !strings.Contains(ReleaseUrl, "Narehatechi/Narecord") {
		t.Errorf("ReleaseUrl should prefer Narecord: %s", ReleaseUrl)
	}
	if !strings.Contains(ReleaseListUrl, "Narehatechi/Narecord") {
		t.Errorf("ReleaseListUrl should prefer Narecord: %s", ReleaseListUrl)
	}
	if strings.Contains(ReleaseUrl, "Equicord/Equicord") {
		t.Errorf("ReleaseUrl should not be Equicord: %s", ReleaseUrl)
	}
	if strings.Contains(ReleaseListUrl, "Equicord/Equicord") {
		t.Errorf("ReleaseListUrl should not be Equicord: %s", ReleaseListUrl)
	}

	fallbacks := []string{ReleaseUrlFallback, ReleaseListUrlFallback, DesktopAsarFallbackUrl}
	for _, u := range fallbacks {
		if !strings.Contains(u, "Equicord/Equicord") {
			t.Errorf("asar fallback should point at Equicord/Equicord: %s", u)
		}
		if strings.Contains(u, "Narehatechi/Narecord") {
			t.Errorf("asar fallback should not be Narecord: %s", u)
		}
		if strings.Contains(u, "releases/tags/v1.0.0") {
			t.Errorf("desktop.asar fallback must not use a missing Narecord v1.0.0 tag: %s", u)
		}
	}
}

func TestInstallerDownloadLinkStaysOnNarecord(t *testing.T) {
	url := GetInstallerDownloadLink()
	if url == "" {
		t.Skip("no installer asset for this OS/arch")
	}
	if !strings.Contains(url, "Narehatechi/Narecord") {
		t.Fatalf("installer download should stay on Narecord: %s", url)
	}
	if strings.Contains(url, "Equicord/Equicord") {
		t.Fatalf("installer download should not use Equicord asar repo: %s", url)
	}
}

func TestInstallerReleaseURLsStayOnNarecord(t *testing.T) {
	urls := []string{
		InstallerReleaseUrl,
		InstallerReleaseUrlFallback,
		UserAgent,
	}
	for _, u := range urls {
		if strings.Contains(u, "blackperson12121") {
			t.Errorf("leftover GitHub owner in %s", u)
		}
		if strings.Contains(u, "Narehatechi/Narelotl") {
			t.Errorf("leftover Narelotl repo path in %s", u)
		}
		if !strings.Contains(u, "Narehatechi/Narecord") {
			t.Errorf("installer URL should stay on Narehatechi/Narecord: %s", u)
		}
		if strings.Contains(u, "Equicord/Equicord") {
			t.Errorf("installer self-update should not point at Equicord: %s", u)
		}
	}
}

func TestDesktopAsarListMaxPagesIsSmall(t *testing.T) {
	if desktopAsarListMaxPages < 1 || desktopAsarListMaxPages > 2 {
		t.Fatalf("desktopAsarListMaxPages = %d, want 1 or 2", desktopAsarListMaxPages)
	}
}

func TestFetchReleaseWithDesktopAsarUsesProcessCache(t *testing.T) {
	resetDesktopAsarReleaseCache()
	t.Cleanup(resetDesktopAsarReleaseCache)

	want := &GithubRelease{
		TagName: "cached",
		Assets: []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		}{
			{Name: "desktop.asar", DownloadURL: "https://example.test/cached.asar"},
		},
	}
	desktopAsarCacheMu.Lock()
	desktopAsarReleaseCache = want
	desktopAsarCacheMu.Unlock()

	got, err := fetchReleaseWithDesktopAsar()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("expected cached release pointer, got %+v", got)
	}
}

func TestInstalledHashRe(t *testing.T) {
	for _, line := range []string{
		"// Narecord abcdef1",
		"// Equicord abcdef1",
		"// Vencord abcdef1",
	} {
		m := installedHashRe.FindSubmatch([]byte(line))
		if m == nil || string(m[1]) != "abcdef1" {
			t.Fatalf("%q -> %q", line, m)
		}
	}
}

func TestIsGithubRateLimitStatus(t *testing.T) {
	for _, s := range []string{"401 Unauthorized", "403 Forbidden", "429 Too Many Requests"} {
		if !isGithubRateLimitStatus(errors.New(s)) {
			t.Fatalf("should treat %q as a rate-limit/block", s)
		}
	}
	if isGithubRateLimitStatus(errors.New("500 Internal Server Error")) {
		t.Fatal("500 is not a rate-limit backoff")
	}
}

func TestFetchReleaseWithDesktopAsarFromNarecordOrEquicord(t *testing.T) {
	resetDesktopAsarReleaseCache()
	t.Cleanup(resetDesktopAsarReleaseCache)
	rel, err := fetchReleaseWithDesktopAsar()
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "403") || strings.Contains(msg, "429") || strings.Contains(msg, "401") {
			t.Skip(err)
		}
		t.Fatalf("should find desktop.asar on Narecord or Equicord fallback: %v", err)
	}
	url := desktopAsarDownloadURL(rel)
	if url == "" {
		t.Fatal("picked a release without desktop.asar")
	}
	if !strings.Contains(url, "desktop.asar") {
		t.Fatalf("url = %s", url)
	}
	switch {
	case strings.Contains(url, "Narehatechi/Narecord"):
		t.Logf("picked Narecord asar %s %s", rel.TagName, url)
	case strings.Contains(url, "Equicord/Equicord"):
		t.Logf("fell back to Equicord asar %s %s", rel.TagName, url)
	default:
		t.Fatalf("expected Narecord or Equicord asar url, got %s", url)
	}
}
