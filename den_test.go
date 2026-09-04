/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	path "path/filepath"
	"runtime"
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
		"Narecord Toolbox",
		"vc-toolbox-btn",
		"Narecord Plugin",
		`alt="User"`,
		"userplugin.png",
		"vc-addon-card",
	} {
		if !strings.Contains(css, needle) {
			t.Errorf("den CSS missing %q", needle)
		}
	}
	if strings.Count(css, "data:image/webp;base64,") < 5 {
		t.Fatalf("expected 5 embedded portraits (old + circle), got %d", strings.Count(css, "data:image/webp;base64,"))
	}
}

func TestDenCSSToolboxAndSettingsUseCirclePortraitsNotPluginCards(t *testing.T) {
	css := denCSS()
	oldURI := dataURI("image/webp", denNarehateWebp)
	oldLgURI := dataURI("image/webp", denNarehateLgWebp)
	circleURI := dataURI("image/webp", denNarehateCircleWebp)
	circleLgURI := dataURI("image/webp", denNarehateCircleLgWebp)

	for _, leftover := range []string{
		"__NAREHATE_CIRCLE_LG__",
		"__NAREHATE_CIRCLE__",
		"__NAREHATE_LG__",
		"__NAREHATE__",
		"__MITTY__",
	} {
		if strings.Contains(css, leftover) {
			t.Errorf("den CSS still has placeholder %q", leftover)
		}
	}

	split := strings.Index(css, "/* narecord-plugin-cards:")
	if split < 0 {
		t.Fatal("missing plugin-cards marker")
	}
	chrome, cards := css[:split], css[split:]

	if !strings.Contains(chrome, circleURI) {
		t.Fatal("toolbox/settings chrome must use narehate-circle.webp")
	}
	if !strings.Contains(chrome, circleLgURI) {
		t.Fatal("toolbox header / den banner must use narehate-circle-lg.webp")
	}
	if strings.Contains(chrome, oldURI) || strings.Contains(chrome, oldLgURI) {
		t.Fatal("toolbox/settings chrome must not keep the old narehate portraits")
	}

	if !strings.Contains(cards, oldURI) {
		t.Fatal("plugin cards must keep the old narehate.webp")
	}
	if !strings.Contains(cards, oldLgURI) {
		t.Fatal("Hideout card must keep the old narehate-lg.webp")
	}
	if strings.Contains(cards, circleURI) || strings.Contains(cards, circleLgURI) {
		t.Fatal("plugin cards must not use the new circle portraits")
	}

	if !bytes.Equal(denNarehateWebp, mustReadFile(t, "assets/den/narehate.webp")) {
		t.Fatal("narehate.webp bytes changed")
	}
	if !bytes.Equal(denNarehateLgWebp, mustReadFile(t, "assets/den/narehate-lg.webp")) {
		t.Fatal("narehate-lg.webp bytes changed")
	}
}

func mustReadFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestDenCSSLeavesEquicordPluginOriginAlone(t *testing.T) {
	css := denCSS()
	for _, needle := range []string{"Show Equicord", "Equicord Plugin", "Equicloud", "vc-plugin-badge", `[alt="Equicord"]`, `[alt="Vencord"]`} {
		if strings.Contains(css, needle) {
			t.Errorf("den CSS should not restyle Equicord/Vencord origin UI %q", needle)
		}
	}
}

func TestDenCSSRetitlesNarecordUserpluginCards(t *testing.T) {
	css := denCSS()
	for _, needle := range []string{
		"Narecord Plugin",
		`alt="User"`,
		"userplugin.png",
		"Layers, relics",
		"Field notebook",
		"repeating-linear-gradient",
		"Thin Abyss depth strip",
		"Nanachi hideout look.",
		"Charged incinerator rail",
		"Small Mitty flair",
		"Quiet den: strips Discord fluff",
		"pointer-events: none",
	} {
		if !strings.Contains(css, needle) {
			t.Errorf("den CSS missing Narecord plugin-card hook %q", needle)
		}
	}
	if strings.Count(css, "content: \"Narecord Plugin\"") < 1 {
		t.Fatal("expected CSS content label Narecord Plugin")
	}
}

func TestDenCSSDoesNotApplyOneDenCoatToEveryPluginCard(t *testing.T) {
	css := denCSS()
	begin := strings.Index(css, "/* narecord-plugin-cards: shared badge/label")
	end := strings.Index(css, "/* narecord-plugin-cards: per-plugin looks */")
	if begin < 0 || end < 0 || end <= begin {
		t.Fatal("missing shared vs per-plugin markers; unique looks cannot be checked")
	}
	shared := css[begin:end]
	unique := css[end:]
	if !strings.Contains(shared, `alt="User"`) || !strings.Contains(shared, "Narecord Plugin") {
		t.Fatal("shared userplugin block should kill the puzzle and label Narecord Plugin")
	}
	if strings.Contains(shared, "linear-gradient(135deg, rgb(108 122 86") {
		t.Fatal("shared userplugin card rule must not paint every card with den moss/rose")
	}
	if strings.Contains(shared, "url(\"data:image/webp") {
		t.Fatal("shared userplugin card rule must not stamp Narehate/Mitty on every card")
	}
	if !strings.Contains(unique, "#7ec8e3") {
		t.Fatal("Abyss card should use layer cyan, not den moss")
	}
	if !strings.Contains(unique, "#f4ead8") || !strings.Contains(unique, "#c45c4a") {
		t.Fatal("NareNotes card should look like a notebook (cream paper + spine)")
	}
	if !strings.Contains(unique, "Nanachi hideout look.") {
		t.Fatal("Hideout is the card that may use den portraits")
	}
	if !strings.Contains(unique, "Quiet den: strips Discord fluff") {
		t.Fatal("narePerf card hook missing from per-plugin looks")
	}
	if !strings.Contains(unique, "clip-path: polygon(0 0, 100% 50%, 0 100%)") {
		t.Fatal("narePerf card must use an angular chevron, not a puzzle/circle sticker")
	}
}

func TestDenCSSDoesNotAsarSwapUserPluginString(t *testing.T) {
	if len("User Plugin") == len("Narecord Plugin") {
		t.Fatal("those labels are not equal length; do not asar-swap them")
	}
	src, err := os.ReadFile("den.go")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(src, []byte("User Plugin")) {
		t.Fatal("den.go must not asar-swap User Plugin; CSS/content is the inject")
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
	tb, _ := plugins[denToolboxPlugin].(map[string]any)
	if tb["enabled"] != true {
		t.Fatal("EquicordToolbox should be enabled so the title-bar toolbox ships")
	}
	np, _ := plugins[narePerfPluginName].(map[string]any)
	if np["enabled"] != true {
		t.Fatal("narePerf should be enabled so the den ships the perf plugin")
	}
	for _, name := range denUserpluginNames {
		p, _ := plugins[name].(map[string]any)
		if p["enabled"] != true {
			t.Fatalf("%s should be enabled so Install ships the den userplugin", name)
		}
	}
	if data["windowsMaterial"] != "none" {
		t.Fatal("windowsMaterial should be none so acrylic/vibrancy is stripped")
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

func manyPluginSettings(n int) map[string]any {
	plugins := make(map[string]any, n)
	for i := 0; i < n; i++ {
		plugins[fmt.Sprintf("Plugin%03d", i)] = map[string]any{
			"enabled": true,
			"note":    "keep-me",
		}
	}
	return map[string]any{
		"useQuickCss":     false,
		"windowsMaterial": "acrylic",
		"plugins":         plugins,
		"bigInt":          json.Number("9007199254740993"),
	}
}

func mustSettingsJSON(t *testing.T, data map[string]any) []byte {
	t.Helper()
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	return append(out, '\n')
}

func parseTestSettings(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	data, err := parseSettingsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestMergeDenSettingsJSONNeverReplacesHundredPluginsWithThree(t *testing.T) {
	raw := mustSettingsJSON(t, manyPluginSettings(100))
	out, err := mergeDenSettingsJSON(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := parseTestSettings(t, out)
	plugins, ok := data["plugins"].(map[string]any)
	if !ok {
		t.Fatal("plugins map missing")
	}
	if len(plugins) < 100 {
		t.Fatalf("plugins wiped: had 100, got %d (the v1.1.6 overwrite bug)", len(plugins))
	}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("Plugin%03d", i)
		p, _ := plugins[name].(map[string]any)
		if p == nil || p["note"] != "keep-me" {
			t.Fatalf("plugin %s was dropped or replaced: %#v", name, p)
		}
	}
	for _, name := range []string{denToolboxPlugin, narePerfPluginName, bundledNoTypingPlugin} {
		p, _ := plugins[name].(map[string]any)
		if p["enabled"] != true {
			t.Fatalf("%s should be enabled without replacing the plugins map", name)
		}
	}
	if data["windowsMaterial"] != "none" {
		t.Fatal("windowsMaterial should be merged to none")
	}
	if data["bigInt"] != json.Number("9007199254740993") {
		t.Fatalf("large integer rewritten: %#v", data["bigInt"])
	}
}

func TestMergeDenSettingsJSONSeedsMissingFromEquicord(t *testing.T) {
	seed := mustSettingsJSON(t, manyPluginSettings(100))
	out, err := mergeDenSettingsJSON(nil, seed)
	if err != nil {
		t.Fatal(err)
	}
	data := parseTestSettings(t, out)
	plugins, _ := data["plugins"].(map[string]any)
	if len(plugins) < 100 {
		t.Fatalf("missing Narecord settings were not seeded from Equicord: %d plugins", len(plugins))
	}
	if plugins["Plugin000"] == nil {
		t.Fatal("seeded Equicord plugin missing")
	}
	np, _ := plugins[narePerfPluginName].(map[string]any)
	if np["enabled"] != true {
		t.Fatal("narePerf should still be enabled after seed")
	}
}

func TestMergeDenSettingsJSONSeedsV11StubFromEquicord(t *testing.T) {
	stub := mustSettingsJSON(t, map[string]any{
		"useQuickCss":     true,
		"windowsMaterial": "none",
		"plugins": map[string]any{
			narePerfPluginName:    map[string]any{"enabled": true},
			bundledNoTypingPlugin: map[string]any{"enabled": true},
			denToolboxPlugin:      map[string]any{"enabled": true},
		},
	})
	if !isSparseInstallerSettings(parseTestSettings(t, stub)) {
		t.Fatal("v1.1.6 3-plugin file should be treated as an installer stub")
	}
	seed := mustSettingsJSON(t, manyPluginSettings(100))
	out, err := mergeDenSettingsJSON(stub, seed)
	if err != nil {
		t.Fatal(err)
	}
	data := parseTestSettings(t, out)
	plugins, _ := data["plugins"].(map[string]any)
	if len(plugins) < 100 {
		t.Fatalf("stub was not replaced by Equicord seed: %d plugins", len(plugins))
	}
	if plugins["Plugin042"] == nil {
		t.Fatal("Equicord plugin 042 missing after stub recovery")
	}
}

func TestMergeDenSettingsJSONRefusesUnparseableAndNonObjectPlugins(t *testing.T) {
	if _, err := mergeDenSettingsJSON([]byte("{ this is not json"), nil); err == nil {
		t.Fatal("unparseable settings must not be replaced with a stub")
	}
	if _, err := mergeDenSettingsJSON([]byte(`{"plugins":["Nope"]}`), nil); err == nil {
		t.Fatal("non-object plugins must not be replaced with a 3-plugin map")
	}
}

func TestInstallDenIntoBacksUpAndPreservesExistingSettings(t *testing.T) {
	root := t.TempDir()
	settingsDir := path.Join(root, "settings")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}
	raw := mustSettingsJSON(t, manyPluginSettings(100))
	settingsPath := path.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, raw, 0644); err != nil {
		t.Fatal(err)
	}

	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}

	bak, err := os.ReadFile(path.Join(settingsDir, settingsBackupName))
	if err != nil {
		t.Fatal("missing settings.json.bak-before-nareperf")
	}
	if !bytes.Equal(bak, raw) {
		t.Fatal("backup is not the original settings.json")
	}

	got, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	data := parseTestSettings(t, got)
	plugins, _ := data["plugins"].(map[string]any)
	if len(plugins) < 100 {
		t.Fatalf("installDenInto overwrote plugins: %d", len(plugins))
	}

	// Second write must not clobber the first backup.
	if err := os.WriteFile(path.Join(settingsDir, settingsBackupName), []byte("first-backup"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}
	bak2, err := os.ReadFile(path.Join(settingsDir, settingsBackupName))
	if err != nil {
		t.Fatal(err)
	}
	if string(bak2) != "first-backup" {
		t.Fatal("reinstall overwrote the first settings backup")
	}
}

func TestInstallDenIntoSeedsFromSiblingAndDoesNotWipeEquicord(t *testing.T) {
	equicord := t.TempDir()
	narecord := t.TempDir()
	if err := os.MkdirAll(path.Join(equicord, "settings"), 0755); err != nil {
		t.Fatal(err)
	}
	seed := mustSettingsJSON(t, manyPluginSettings(100))
	if err := os.WriteFile(path.Join(equicord, "settings", "settings.json"), seed, 0644); err != nil {
		t.Fatal(err)
	}

	picked := richestSettingsJSON([]string{narecord, equicord})
	if settingsPluginCount(parseTestSettings(t, picked)) != 100 {
		t.Fatal("richestSettingsJSON should pick Equicord's 100-plugin file")
	}

	if err := installDenIntoWithSeed(narecord, picked); err != nil {
		t.Fatal(err)
	}
	if err := installDenIntoWithSeed(equicord, picked); err != nil {
		t.Fatal(err)
	}

	narecordSettings := parseTestSettings(t, mustReadFile(t, path.Join(narecord, "settings", "settings.json")))
	equicordSettings := parseTestSettings(t, mustReadFile(t, path.Join(equicord, "settings", "settings.json")))
	if settingsPluginCount(narecordSettings) < 100 {
		t.Fatalf("Narecord was created as a 3-plugin stub: %d", settingsPluginCount(narecordSettings))
	}
	if settingsPluginCount(equicordSettings) < 100 {
		t.Fatalf("Equicord plugins were wiped: %d", settingsPluginCount(equicordSettings))
	}
	eqPlugins, _ := equicordSettings["plugins"].(map[string]any)
	p, _ := eqPlugins["Plugin099"].(map[string]any)
	if p["note"] != "keep-me" {
		t.Fatal("Equicord plugin settings were replaced, not merged")
	}
}

func TestInstallDenIntoLeavesUnparseableSettingsUntouched(t *testing.T) {
	root := t.TempDir()
	settingsDir := path.Join(root, "settings")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := path.Join(settingsDir, "settings.json")
	garbage := []byte("{ this is not json")
	if err := os.WriteFile(settingsPath, garbage, 0644); err != nil {
		t.Fatal(err)
	}
	if err := installDenInto(root); err == nil {
		t.Fatal("expected error for unparseable settings.json")
	}
	got := mustReadFile(t, settingsPath)
	if !bytes.Equal(got, garbage) {
		t.Fatalf("unparseable settings.json was overwritten: %s", got)
	}
}

func TestInstallDenIntoPreservesUserQuickCSSBlocks(t *testing.T) {
	root := t.TempDir()
	settingsDir := path.Join(root, "settings")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatal(err)
	}
	user := "/* my custom css */\n.foo { color: red; }\n"
	if err := os.WriteFile(path.Join(settingsDir, "quickCss.css"), []byte(user), 0644); err != nil {
		t.Fatal(err)
	}
	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}
	got := string(mustReadFile(t, path.Join(settingsDir, "quickCss.css")))
	if !strings.Contains(got, "my custom css") || !strings.Contains(got, ".foo { color: red; }") {
		t.Fatalf("user QuickCSS was wiped: %s", got)
	}
	if strings.Count(got, denQuickCSSBegin) != 1 || strings.Count(got, denQuickCSSEnd) != 1 {
		t.Fatal("den QuickCSS markers missing or duplicated")
	}
	if strings.Count(got, narePerfQuickCSSBegin) != 1 || strings.Count(got, narePerfQuickCSSEnd) != 1 {
		t.Fatal("narePerf QuickCSS markers missing or duplicated")
	}
	if err := installDenInto(root); err != nil {
		t.Fatal(err)
	}
	again := string(mustReadFile(t, path.Join(settingsDir, "quickCss.css")))
	if strings.Count(again, denQuickCSSBegin) != 1 || strings.Count(again, narePerfQuickCSSBegin) != 1 {
		t.Fatal("reinstall duplicated QuickCSS marker blocks")
	}
	if !strings.Contains(again, "my custom css") {
		t.Fatal("user QuickCSS lost on reinstall")
	}
}

func TestRichestSettingsJSONIgnoresUnreadableFiles(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	if err := os.MkdirAll(path.Join(a, "settings"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(a, "settings", "settings.json"), []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}
	if richestSettingsJSON([]string{a, b}) != nil {
		t.Fatal("unreadable settings should not be chosen as seed")
	}
}

func TestIsSparseInstallerSettings(t *testing.T) {
	if !isSparseInstallerSettings(nil) || !isSparseInstallerSettings(map[string]any{}) {
		t.Fatal("empty settings are sparse")
	}
	if isSparseInstallerSettings(manyPluginSettings(4)) {
		t.Fatal("user plugins must not look like the installer stub")
	}
	if !isSparseInstallerSettings(map[string]any{
		"plugins": map[string]any{narePerfPluginName: map[string]any{"enabled": true}},
	}) {
		t.Fatal("single installer plugin is still a stub")
	}
}

func TestEnsurePluginsMapRefusesNonObject(t *testing.T) {
	data := map[string]any{"plugins": []any{"x"}}
	if _, err := ensurePluginsMap(data); err == nil {
		t.Fatal("expected refuse")
	}
	if _, ok := data["plugins"].([]any); !ok {
		t.Fatal("non-object plugins value was replaced")
	}
}

func TestParseSettingsJSONStripsBOMAndPreservesNumbers(t *testing.T) {
	raw := append(append([]byte{}, utf8BOM...), []byte(`{"n": 9007199254740993, "plugins": {}}`)...)
	data, err := parseSettingsJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if data["n"] != json.Number("9007199254740993") {
		t.Fatalf("number = %#v", data["n"])
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
	if !strings.Contains(string(theme), "Narecord Plugin") || !strings.Contains(string(theme), "Field notebook") {
		t.Fatal("theme missing Narecord plugin cards")
	}

	quick, err := os.ReadFile(path.Join(root, "settings", "quickCss.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(quick), denQuickCSSBegin) || !strings.Contains(string(quick), "equicord_main") {
		t.Fatalf("quickCss missing den: %s", quick)
	}
	if !strings.Contains(string(quick), "Narecord Toolbox") || !strings.Contains(string(quick), "vc-toolbox-btn") {
		t.Fatalf("quickCss missing toolbox den: %s", quick)
	}
	if !strings.Contains(string(quick), "Narecord Plugin") || !strings.Contains(string(quick), `alt="User"`) {
		t.Fatalf("quickCss missing Narecord plugin cards: %s", quick)
	}
	if !strings.Contains(string(quick), "Field notebook") || !strings.Contains(string(quick), "Layers, relics") {
		t.Fatalf("quickCss missing per-plugin card looks: %s", quick)
	}

	settings, err := os.ReadFile(path.Join(root, "settings", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(settings), denThemeFileName) || !strings.Contains(string(settings), `"useQuickCss": true`) {
		t.Fatalf("settings.json not enabled: %s", settings)
	}
	if !strings.Contains(string(settings), denToolboxPlugin) {
		t.Fatalf("settings.json did not enable toolbox plugin: %s", settings)
	}
	if !strings.Contains(string(settings), narePerfPluginName) || !strings.Contains(string(quick), narePerfQuickCSSBegin) {
		t.Fatalf("narePerf was not shipped: settings=%s quick=%s", settings, quick)
	}
	if !strings.Contains(string(quick), "html:not(.nr-perf-plugin)") || !strings.Contains(string(quick), "backdrop-filter: none") {
		t.Fatalf("quickCss missing narePerf fluff-kill fallback: %s", quick)
	}
	pluginSrc, err := os.ReadFile(path.Join(root, "userplugins", narePerfPluginName, narePerfPluginFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(pluginSrc), "definePlugin") || !strings.Contains(string(pluginSrc), narePerfCardHook) {
		t.Fatalf("userplugin source incomplete: %s", pluginSrc)
	}
	abyss, err := os.ReadFile(path.Join(root, "userplugins", "Abyss", "index.tsx"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(abyss), "definePlugin") || !strings.Contains(string(abyss), `name: "Abyss"`) {
		t.Fatalf("Abyss userplugin missing: %s", abyss)
	}
	if !ExistsFile(path.Join(root, "userplugins", "Hideout", "style.css")) {
		t.Fatal("Hideout style.css was not written")
	}
}

func TestPatchAsarEquicordToolboxRetitlesWithoutTouchingPluginId(t *testing.T) {
	if len(equicordToolboxLabel) != len(narecordToolboxLabel) {
		t.Fatalf("labels must be equal length: %d vs %d", len(equicordToolboxLabel), len(narecordToolboxLabel))
	}
	dir := t.TempDir()
	asar := path.Join(dir, "desktop.asar")
	in := []byte("id:EquicordToolbox tooltip:Equicord Toolbox more Equicord Toolbox end")
	if err := os.WriteFile(asar, in, 0644); err != nil {
		t.Fatal(err)
	}
	if err := patchAsarEquicordToolbox(asar); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(asar)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if strings.Contains(got, "Equicord Toolbox") {
		t.Fatalf("label left behind: %s", got)
	}
	if !strings.Contains(got, "Narecord Toolbox") {
		t.Fatalf("Narecord Toolbox missing: %s", got)
	}
	if strings.Count(got, "Narecord Toolbox") != 2 {
		t.Fatalf("expected 2 Narecord Toolbox, got %s", got)
	}
	if !strings.Contains(got, "EquicordToolbox") {
		t.Fatal("plugin id EquicordToolbox must stay")
	}
}

func TestPatchAsarEquicordToolboxNoopWhenAlreadyNarecord(t *testing.T) {
	dir := t.TempDir()
	asar := path.Join(dir, "desktop.asar")
	in := []byte("tooltip:Narecord Toolbox plugin:EquicordToolbox")
	if err := os.WriteFile(asar, in, 0644); err != nil {
		t.Fatal(err)
	}
	if err := patchAsarEquicordToolbox(asar); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(asar)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(in) {
		t.Fatalf("asar changed when already Narecord: %s", out)
	}
}

func TestReplaceAllEqualLengthInPlaceSpansChunks(t *testing.T) {
	dir := t.TempDir()
	asar := path.Join(dir, "desktop.asar")
	// 7-byte prefix so with 8-byte chunks the first needle starts 1 byte before a
	// boundary; the second sits after a 3-byte gap.
	in := append([]byte("xxxxxxx"), equicordToolboxLabel...)
	in = append(in, []byte("yyy")...)
	in = append(in, equicordToolboxLabel...)
	in = append(in, []byte("zz")...)
	if err := os.WriteFile(asar, in, 0644); err != nil {
		t.Fatal(err)
	}
	n, err := replaceAllEqualLengthInPlaceChunked(asar, equicordToolboxLabel, narecordToolboxLabel, 8)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("replaced %d, want 2", n)
	}
	out, err := os.ReadFile(asar)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, equicordToolboxLabel) {
		t.Fatalf("needle left behind: %s", out)
	}
	if bytes.Count(out, narecordToolboxLabel) != 2 {
		t.Fatalf("expected 2 replacements, got %s", out)
	}
	want := append([]byte("xxxxxxx"), narecordToolboxLabel...)
	want = append(want, []byte("yyy")...)
	want = append(want, narecordToolboxLabel...)
	want = append(want, []byte("zz")...)
	if !bytes.Equal(out, want) {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestReplaceAllEqualLengthInPlacePreservesModeAndSurroundingBytes(t *testing.T) {
	dir := t.TempDir()
	asar := path.Join(dir, "desktop.asar")
	in := append(bytes.Repeat([]byte{0xAA}, 64), equicordToolboxLabel...)
	in = append(in, bytes.Repeat([]byte{0xBB}, 64)...)
	if err := os.WriteFile(asar, in, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := replaceAllEqualLengthInPlace(asar, equicordToolboxLabel, narecordToolboxLabel); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(asar)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out[:64], bytes.Repeat([]byte{0xAA}, 64)) || !bytes.Equal(out[len(out)-64:], bytes.Repeat([]byte{0xBB}, 64)) {
		t.Fatal("bytes outside the label were rewritten")
	}
	if runtime.GOOS != "windows" {
		st, err := os.Stat(asar)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0600 {
			t.Fatalf("mode = %o, want 0600", st.Mode().Perm())
		}
	}
}

func TestFindBytesOffsetsMatchesReplaceAll(t *testing.T) {
	needle := []byte("Equicord Toolbox")
	cases := [][]byte{
		{},
		[]byte("nope"),
		needle,
		append(needle, needle...),
		append([]byte("a"), append(needle, 'b')...),
		append(bytes.Repeat([]byte("z"), 100), needle...),
		append(needle, bytes.Repeat([]byte("z"), 100)...),
	}
	for i, in := range cases {
		offs, err := findBytesOffsets(bytes.NewReader(in), needle, 9)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		want := bytes.Count(in, needle)
		if len(offs) != want {
			t.Fatalf("case %d: got %d offsets, want %d", i, len(offs), want)
		}
		for _, off := range offs {
			if off < 0 || int(off)+len(needle) > len(in) {
				t.Fatalf("case %d: offset %d out of range", i, off)
			}
			if !bytes.Equal(in[off:int(off)+len(needle)], needle) {
				t.Fatalf("case %d: offset %d is not a match", i, off)
			}
		}
	}
}

func TestDenCSSIsCached(t *testing.T) {
	a := denCSS()
	b := denCSS()
	if a != b {
		t.Fatal("cached den CSS should be stable")
	}
	if denCSSCached == "" || denCSSCached != a {
		t.Fatal("denCSSOnce did not populate denCSSCached")
	}
}
