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

func TestNarePerfPluginIsRealVencordPlugin(t *testing.T) {
	if err := narePerfPluginValid(); err != nil {
		t.Fatal(err)
	}
	src := narePerfPluginSource
	for _, needle := range []string{
		"export default definePlugin",
		`name: "narePerf"`,
		"definePluginSettings",
		"killBlur",
		"killAnimations",
		"killAcrylic",
		"killDecorations",
		"quietTyping",
		"quietGifs",
		"skipSprings",
		"start()",
		"stop()",
		"dotCycle",
		"canAnimate:",
		"messageContent",
	} {
		if !strings.Contains(src, needle) {
			t.Errorf("plugin missing %q", needle)
		}
	}
	for _, banned := range []string{"sendMessage", "RTCConnection", "streamKey", "eval("} {
		if strings.Contains(src, banned) {
			t.Errorf("plugin must not contain %q", banned)
		}
	}
}

func TestNarePerfCardLookIsDistinctInstrumentNotPuzzle(t *testing.T) {
	css := denCSS()
	hook := `.vc-addon-card:has(.vc-plugins-source[alt="User"]):has(.vc-addon-note[title^="Quiet den: strips Discord fluff"])`
	idx := strings.Index(css, hook)
	if idx < 0 {
		t.Fatal("narePerf card selector missing")
	}
	next := strings.Index(css[idx+len(hook):], "/* ")
	block := css[idx:]
	if next > 0 {
		block = css[idx : idx+len(hook)+next]
	}
	for _, needle := range []string{
		"clip-path: polygon(0 0, 100% 50%, 0 100%)",
		"#8fbf4a",
		"#161a16",
		"border-radius: 2px",
		"box-shadow: none",
	} {
		if !strings.Contains(block, needle) {
			t.Errorf("narePerf card missing %q", needle)
		}
	}
	for _, banned := range []string{
		"userplugin.png",
		"url(\"data:image/webp",
		"puzzle",
		"#1e6eff",
		"__NAREHATE__",
		"__MITTY__",
	} {
		if strings.Contains(block, banned) {
			t.Errorf("narePerf card must not use %q", banned)
		}
	}
}

func TestNarePerfFallbackYieldsToCompiledPlugin(t *testing.T) {
	if !strings.Contains(narePerfFallbackCSS, "html:not(.nr-perf-plugin)") {
		t.Fatal("fallback must not run while the plugin owns html.nr-perf-plugin")
	}
	if !strings.Contains(narePerfFallbackCSS, "backdrop-filter: none") {
		t.Fatal("fallback must kill blur")
	}
	if !strings.Contains(narePerfFallbackCSS, "animation-duration: 0.001ms") {
		t.Fatal("fallback must use reduced-motion durations, not animation:none on chat")
	}
	if !strings.Contains(narePerfFallbackCSS, "messageContent") {
		t.Fatal("fallback must spare message text selectors")
	}
}

func TestWriteNarePerfUserpluginAndSourceTree(t *testing.T) {
	root := t.TempDir()
	if err := writeNarePerfUserplugin(root); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path.Join(root, "userplugins", "narePerf", "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != narePerfPluginSource {
		t.Fatal("written userplugin does not match embedded source")
	}

	srcRoot := t.TempDir()
	if err := os.MkdirAll(path.Join(srcRoot, "src", "equicordplugins"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeNarePerfIntoSourceTree(srcRoot); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(path.Join(srcRoot, "src", "userplugins", "narePerf", "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `name: "narePerf"`) {
		t.Fatal("source-tree inject missing plugin")
	}

	plain := t.TempDir()
	if err := writeNarePerfIntoSourceTree(plain); err != nil {
		t.Fatal(err)
	}
	if ExistsFile(path.Join(plain, "src", "userplugins", "narePerf", "index.ts")) {
		t.Fatal("must not write into a random directory that is not Equicord source")
	}
}

func TestEnableNarePerfInSettingsJSON(t *testing.T) {
	data := map[string]any{
		"plugins": map[string]any{
			"Foo": map[string]any{"enabled": true},
		},
		"windowsMaterial": "acrylic",
	}
	enableNarePerfInSettings(data)
	if data["windowsMaterial"] != "none" {
		t.Fatal("acrylic must be stripped")
	}
	plugins, _ := data["plugins"].(map[string]any)
	if plugins["Foo"] == nil {
		t.Fatal("other plugins wiped")
	}
	for _, name := range []string{narePerfPluginName, bundledNoTypingPlugin} {
		p, _ := plugins[name].(map[string]any)
		if p["enabled"] != true {
			t.Fatalf("%s should be enabled", name)
		}
	}
}

func TestMergeNarePerfQuickCSSReplacesBlock(t *testing.T) {
	first := mergeNarePerfQuickCSS("/* user */")
	if !strings.Contains(first, "/* user */") || !strings.Contains(first, narePerfQuickCSSBegin) {
		t.Fatalf("expected user CSS and narePerf block, got %s", first)
	}
	second := mergeNarePerfQuickCSS(first)
	if strings.Count(second, narePerfQuickCSSBegin) != 1 {
		t.Fatal("expected a single narePerf block")
	}
}

func TestLooksLikeEquicordSource(t *testing.T) {
	root := t.TempDir()
	if looksLikeEquicordSource(root) {
		t.Fatal("empty dir is not Equicord source")
	}
	if err := os.WriteFile(path.Join(root, "src"), []byte("no"), 0644); err != nil {
		t.Fatal(err)
	}
	if looksLikeEquicordSource(root) {
		t.Fatal("a file named src is not Equicord source")
	}
	src := t.TempDir()
	if err := os.MkdirAll(path.Join(src, "src", "plugins"), 0755); err != nil {
		t.Fatal(err)
	}
	if !looksLikeEquicordSource(src) {
		t.Fatal("src/plugins should count as Equicord source")
	}
}

func TestInstallDenShipsNarePerfEndToEnd(t *testing.T) {
	root := t.TempDir()
	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}
	settings, err := os.ReadFile(path.Join(root, "settings", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(settings, &data); err != nil {
		t.Fatal(err)
	}
	if data["windowsMaterial"] != "none" {
		t.Fatalf("windowsMaterial = %#v", data["windowsMaterial"])
	}
	plugins, _ := data["plugins"].(map[string]any)
	np, _ := plugins[narePerfPluginName].(map[string]any)
	if np["enabled"] != true {
		t.Fatal("narePerf not enabled")
	}
}
