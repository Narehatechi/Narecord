/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
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

func TestNarehatechiURLs(t *testing.T) {
	urls := []string{
		ReleaseUrl,
		ReleaseUrlFallback,
		ReleaseListUrl,
		ReleaseListUrlFallback,
		DesktopAsarFallbackUrl,
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
		if !strings.Contains(u, "Narehatechi") {
			t.Errorf("expected Narehatechi in %s", u)
		}
	}
}

func TestFetchReleaseWithDesktopAsarWalksInstallerOnlyLatest(t *testing.T) {
	rel, err := fetchReleaseWithDesktopAsar()
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "403") || strings.Contains(msg, "429") || strings.Contains(msg, "401") {
			t.Skip(err)
		}
		t.Fatalf("walk should find desktop.asar: %v", err)
	}
	url := desktopAsarDownloadURL(rel)
	if url == "" {
		t.Fatal("picked a release without desktop.asar")
	}
	if rel.TagName == "v1.1.0" {
		t.Fatalf("picked installer-only latest %s", rel.TagName)
	}
	if !strings.Contains(url, "desktop.asar") {
		t.Fatalf("url = %s", url)
	}
	t.Logf("picked %s %s", rel.TagName, url)
}
