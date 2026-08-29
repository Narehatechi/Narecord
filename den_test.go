/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"encoding/json"
	"os"
	path "path/filepath"
	"strings"
	"testing"
)

func TestDenCSSEmbedsPortraitsAndDenCopy(t *testing.T) {
	css := denCSS()
	for _, needle := range []string{
		"Narecord Settings",
		"Narecord",
		"The Den",
		"data:image/webp;base64,",
		"equicord_main",
		"equicord_plugins",
		"equicord_section",
		"vc-special-card",
	} {
		if !strings.Contains(css, needle) {
			t.Errorf("den CSS missing %q", needle)
		}
	}
	if strings.Count(css, "data:image/webp;base64,") < 3 {
		t.Fatalf("expected 3 embedded portraits, got %d", strings.Count(css, "data:image/webp;base64,"))
	}
}

func TestDenCSSDoesNotTouchPluginOriginUI(t *testing.T) {
	css := denCSS()
	for _, needle := range []string{"Show Equicord", "Equicord Plugin", "Equicloud", "vc-plugin-badge"} {
		if strings.Contains(css, needle) {
			t.Errorf("den CSS should not target plugin-origin UI %q", needle)
		}
	}
}

func TestMergeQuickCSSReplacesMarkedBlock(t *testing.T) {
	first := mergeQuickCSS("/* user css */\nbody{color:red}", "A { color: moss; }")
	if !strings.Contains(first, "/* user css */") || !strings.Contains(first, denQuickCSSBegin) {
		t.Fatalf("expected user CSS and den block, got %s", first)
	}
	second := mergeQuickCSS(first, "B { color: rose; }")
	if strings.Contains(second, "A { color: moss; }") {
		t.Fatal("old den block was not replaced")
	}
	if strings.Count(second, denQuickCSSBegin) != 1 {
		t.Fatal("expected a single den block after reinstall")
	}
	if !strings.Contains(second, "B { color: rose; }") {
		t.Fatal("new den CSS missing")
	}
}

func TestEnableDenInSettingsJSONPreservesOtherKeys(t *testing.T) {
	raw := []byte(`{"useQuickCss":false,"enabledThemes":["keep.me.css"],"plugins":{"Foo":{"enabled":true}}}`)
	out, err := enableDenInSettingsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(out, &data); err != nil {
		t.Fatal(err)
	}
	if data["useQuickCss"] != true {
		t.Fatal("useQuickCss should be enabled so the den QuickCSS loads")
	}
	themes, _ := data["enabledThemes"].([]any)
	got := map[string]bool{}
	for _, tval := range themes {
		got[tval.(string)] = true
	}
	if !got["keep.me.css"] || !got[denThemeFileName] {
		t.Fatalf("enabledThemes = %#v", themes)
	}
	plugins, _ := data["plugins"].(map[string]any)
	if plugins["Foo"] == nil {
		t.Fatal("existing plugin settings were wiped")
	}

	out2, err := enableDenInSettingsJSON(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out2, &data); err != nil {
		t.Fatal(err)
	}
	themes, _ = data["enabledThemes"].([]any)
	count := 0
	for _, tval := range themes {
		if tval.(string) == denThemeFileName {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("theme enabled twice: %#v", themes)
	}
}

func TestInstallDenIntoWritesThemeQuickCSSAndSettings(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(path.Join(root, "pre-existing"), []byte("no"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}

	theme, err := os.ReadFile(path.Join(root, "themes", denThemeFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(theme), "@name Narecord Den") {
		t.Fatal("theme header missing")
	}
	if !strings.Contains(string(theme), "data:image/webp;base64,") {
		t.Fatal("theme missing portraits")
	}

	quick, err := os.ReadFile(path.Join(root, "settings", "quickCss.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(quick), denQuickCSSBegin) || !strings.Contains(string(quick), "equicord_main") {
		t.Fatalf("quickCss missing den: %s", quick)
	}

	settings, err := os.ReadFile(path.Join(root, "settings", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(settings), denThemeFileName) || !strings.Contains(string(settings), `"useQuickCss": true`) {
		t.Fatalf("settings.json not enabled: %s", settings)
	}
}
