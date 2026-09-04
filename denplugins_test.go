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

func TestDenUserpluginsValid(t *testing.T) {
	if err := denUserpluginsValid(); err != nil {
		t.Fatal(err)
	}
}

func TestWriteDenUserpluginsWritesEveryEmbeddedFile(t *testing.T) {
	root := t.TempDir()
	if err := writeDenUserplugins(root); err != nil {
		t.Fatal(err)
	}
	wantFiles := map[string][]string{
		"Abyss":         {"index.tsx"},
		"Hideout":       {"index.ts", "style.css"},
		"Incinerator":   {"index.ts", "style.css"},
		"NareMotion":    {"index.ts", "style.css"},
		"NarehateBadge": {"index.ts"},
		"Narelogs":      {"index.tsx", "style.css"},
		"NareNotes":     {"index.tsx", "style.css"},
		"Nnaa":          {"index.tsx"},
	}
	for _, name := range denUserpluginNames {
		files, ok := wantFiles[name]
		if !ok {
			t.Fatalf("missing expected-file list for %s", name)
		}
		var src []byte
		var err error
		for _, file := range files {
			p := path.Join(root, "userplugins", name, file)
			got, readErr := os.ReadFile(p)
			if readErr != nil {
				t.Fatalf("missing %s: %v", p, readErr)
			}
			if strings.HasPrefix(file, "index.") {
				src = got
				err = nil
			}
			_ = err
		}
		if !strings.Contains(string(src), "definePlugin") || !strings.Contains(string(src), `name: "`+name+`"`) {
			t.Fatalf("%s index missing definePlugin/name: %s", name, src)
		}
	}
}

func TestWriteDenUserpluginsIntoSourceTree(t *testing.T) {
	srcRoot := t.TempDir()
	if err := os.MkdirAll(path.Join(srcRoot, "src", "equicordplugins"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDenUserpluginsIntoSourceTree(srcRoot); err != nil {
		t.Fatal(err)
	}
	if !ExistsFile(path.Join(srcRoot, "src", "userplugins", "Abyss", "index.tsx")) {
		t.Fatal("expected Abyss in Equicord src/userplugins")
	}
	if !ExistsFile(path.Join(srcRoot, "src", "userplugins", "Hideout", "style.css")) {
		t.Fatal("expected Hideout style.css in source tree")
	}

	plain := t.TempDir()
	if err := writeDenUserpluginsIntoSourceTree(plain); err != nil {
		t.Fatal(err)
	}
	if ExistsFile(path.Join(plain, "src", "userplugins", "Abyss", "index.tsx")) {
		t.Fatal("must not write den plugins into a random directory")
	}
}

func TestEnableDenUserpluginsInSettingsMergesOnly(t *testing.T) {
	original := map[string]any{
		"Foo": map[string]any{"enabled": true, "keep": "yes"},
		"Bar": map[string]any{"enabled": false, "volume": 3},
	}
	data := map[string]any{"plugins": original}
	if err := enableDenUserpluginsInSettings(data); err != nil {
		t.Fatal(err)
	}
	got, _ := data["plugins"].(map[string]any)
	got["__probe"] = true
	if original["__probe"] != true {
		t.Fatal("plugins map was replaced; Install must mutate the existing map")
	}
	if got["Foo"] == nil || got["Bar"] == nil {
		t.Fatal("other plugins wiped")
	}
	foo, _ := got["Foo"].(map[string]any)
	bar, _ := got["Bar"].(map[string]any)
	if foo["keep"] != "yes" || bar["volume"] != 3 || bar["enabled"] != false {
		t.Fatalf("other plugin entries were rewritten: foo=%#v bar=%#v", foo, bar)
	}
	for _, name := range denUserpluginNames {
		p, _ := got[name].(map[string]any)
		if p["enabled"] != true {
			t.Fatalf("%s should be enabled", name)
		}
	}
}

func TestInstallDenShipsDenUserpluginsEndToEnd(t *testing.T) {
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
	for _, name := range denUserpluginNames {
		p, _ := plugins[name].(map[string]any)
		if p["enabled"] != true {
			t.Fatalf("%s not enabled in settings", name)
		}
		if !ExistsFile(path.Join(root, "userplugins", name)) {
			t.Fatalf("userplugins/%s missing", name)
		}
	}
	np, _ := plugins[narePerfPluginName].(map[string]any)
	if np["enabled"] != true {
		t.Fatal("narePerf must still be enabled")
	}
}

func TestDenPluginCardsStayPerPluginNotOneCoat(t *testing.T) {
	css := denCSS()
	if !strings.Contains(css, `title^="Layers, relics"`) {
		t.Fatal("Abyss must keep its own layers card")
	}
	if !strings.Contains(css, `title^="Field notebook"`) {
		t.Fatal("NareNotes must keep its notebook card")
	}
	if !strings.Contains(css, `title="Nanachi hideout look."`) {
		t.Fatal("Hideout must keep its own hideout card")
	}
	src, err := denPluginIndexSource("NareNotes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(src, `name: "NareNotes"`) || !strings.Contains(src, "Field notebook") {
		t.Fatal("NareNotes source must keep the Field notebook description the card CSS keys off")
	}
}

func TestReadmeDocumentsDenAsarBake(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	for _, needle := range []string{
		"build-desktop-asar",
		"Total Userplugins: 0",
		"desktop.asar",
		"falls back to Equicord",
		"Abyss stays layers",
		"NareNotes stays a notebook",
	} {
		if !strings.Contains(doc, needle) {
			t.Errorf("README missing den asar-bake callout %q", needle)
		}
	}
}

func TestPluginsDirHasEveryAsarBakeTarget(t *testing.T) {
	want := append(append([]string{}, denUserpluginNames...), narePerfPluginName)
	entries, err := os.ReadDir("plugins")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			got[e.Name()] = true
		}
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("plugins/%s missing; Release asar bake copies every folder under plugins/", name)
		}
	}
	if len(got) != len(want) {
		t.Errorf("plugins/ has %d folders, asar bake expects %d (%v vs %v)", len(got), len(want), got, want)
	}
}

func TestIsInstallerManagedPluginIncludesDenSet(t *testing.T) {
	for _, name := range denUserpluginNames {
		if !isInstallerManagedPlugin(name) {
			t.Fatalf("%s should be installer-managed for sparse-stub detection", name)
		}
	}
	if isInstallerManagedPlugin("FooUserPlugin") {
		t.Fatal("random user plugin must not look installer-managed")
	}
	if !isSparseInstallerSettings(map[string]any{
		"plugins": map[string]any{
			"Abyss":            map[string]any{"enabled": true},
			narePerfPluginName: map[string]any{"enabled": true},
		},
	}) {
		t.Fatal("only installer-managed plugins is still a stub")
	}
}
