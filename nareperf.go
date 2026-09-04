/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"errors"
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	_ "embed"
)

// narePerf is a first-party den userplugin. Stock Narelotl downloads Equicord's
// desktop.asar, which only lists plugins compiled into that asar. The installer
// therefore:
//  1. writes plugin source to <dataDir>/userplugins/narePerf (always)
//  2. copies it into Equicord src/userplugins when a source tree is found
//  3. enables narePerf + bundled NoTypingAnimation in settings.json
//  4. sets windowsMaterial to none
//  5. injects fallback CSS into QuickCSS so fluff-kill works without a rebuild
const (
	narePerfPluginName     = "narePerf"
	narePerfPluginFile     = "index.ts"
	narePerfUserpluginsRel = "userplugins/narePerf"
	narePerfQuickCSSBegin  = "/* NARECORD-NARPERF-BEGIN */"
	narePerfQuickCSSEnd    = "/* NARECORD-NARPERF-END */"
	narePerfCardHook       = "Quiet den: strips Discord fluff"
	bundledNoTypingPlugin  = "NoTypingAnimation"
)

//go:embed plugins/narePerf/index.ts
var narePerfPluginSource string

//go:embed plugins/narePerf/fallback.css
var narePerfFallbackCSS string

func mergeMarkedCSS(existing, begin, end, body string) string {
	block := begin + "\n" + body + "\n" + end
	start := strings.Index(existing, begin)
	stop := strings.Index(existing, end)
	if start >= 0 && stop > start {
		return existing[:start] + block + existing[stop+len(end):]
	}
	existing = strings.TrimRight(existing, "\n")
	if existing == "" {
		return block + "\n"
	}
	return existing + "\n\n" + block + "\n"
}

func enableNarePerfInSettings(data map[string]any) {
	data["windowsMaterial"] = "none"
	data["macosVibrancyStyle"] = nil

	plugins, _ := data["plugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}

	enable := func(name string) {
		p, _ := plugins[name].(map[string]any)
		if p == nil {
			p = map[string]any{}
		}
		p["enabled"] = true
		plugins[name] = p
	}
	enable(narePerfPluginName)
	enable(bundledNoTypingPlugin)
	data["plugins"] = plugins
}

func narePerfPluginValid() error {
	for _, needle := range []string{
		"definePlugin",
		"export default definePlugin",
		`name: "narePerf"`,
		narePerfCardHook,
		"start()",
		"stop()",
		"killBlur",
		"killAnimations",
		"killAcrylic",
		"definePluginSettings",
	} {
		if !strings.Contains(narePerfPluginSource, needle) {
			return fmt.Errorf("narePerf plugin source missing %q", needle)
		}
	}
	if strings.Contains(narePerfPluginSource, "sendMessage") || strings.Contains(narePerfPluginSource, "RTCConnection") {
		return errors.New("narePerf must not patch message send or voice")
	}
	if !strings.Contains(narePerfFallbackCSS, "html:not(.nr-perf-plugin)") {
		return errors.New("narePerf fallback CSS must yield to the compiled plugin")
	}
	if !strings.Contains(narePerfFallbackCSS, "backdrop-filter: none") {
		return errors.New("narePerf fallback CSS missing blur kill")
	}
	if !strings.Contains(narePerfFallbackCSS, "messageContent") {
		return errors.New("narePerf fallback CSS must spare chat message text")
	}
	return nil
}

func writeNarePerfUserplugin(root string) error {
	dir := path.Join(root, "userplugins", narePerfPluginName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	dst := path.Join(dir, narePerfPluginFile)
	if err := os.WriteFile(dst, []byte(narePerfPluginSource), 0644); err != nil {
		return err
	}
	return nil
}

func looksLikeEquicordSource(root string) bool {
	for _, rel := range []string{
		path.Join("src", "equicordplugins"),
		path.Join("src", "plugins"),
		path.Join("src", "Vencord.ts"),
	} {
		if ExistsFile(path.Join(root, rel)) {
			return true
		}
	}
	return false
}

func writeNarePerfIntoSourceTree(root string) error {
	if root == "" || !looksLikeEquicordSource(root) {
		return nil
	}
	dir := path.Join(root, "src", "userplugins", narePerfPluginName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path.Join(dir, narePerfPluginFile), []byte(narePerfPluginSource), 0644)
}

func narePerfSourceRoots() []string {
	var roots []string
	seen := map[string]bool{}
	add := func(d string) {
		d = strings.TrimSpace(d)
		if d == "" || seen[d] {
			return
		}
		st, err := os.Stat(d)
		if err != nil || !st.IsDir() {
			return
		}
		seen[d] = true
		roots = append(roots, d)
	}
	add(NarecordDirectory)
	add(os.Getenv("EQUICORD_DIRECTORY"))
	add(os.Getenv("NARECORD_DIRECTORY"))
	if NarecordDirectory != "" {
		add(path.Dir(NarecordDirectory))
		add(path.Join(NarecordDirectory, ".."))
	}
	return roots
}

func installNarePerfSources() {
	for _, root := range narePerfSourceRoots() {
		if err := writeNarePerfIntoSourceTree(root); err != nil {
			Log.Warn("Could not write narePerf into Equicord source", root, err)
		} else if looksLikeEquicordSource(root) {
			Log.Info("Wrote narePerf userplugin into Equicord source", root)
		}
	}
}

func mergeNarePerfQuickCSS(existing string) string {
	return mergeMarkedCSS(existing, narePerfQuickCSSBegin, narePerfQuickCSSEnd, strings.TrimSpace(narePerfFallbackCSS))
}
