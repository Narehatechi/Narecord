/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	path "path/filepath"
	"strings"

	"github.com/ProtonMail/go-appdir"

	_ "embed"
)

// Den files written on Install / Reinstall / Repair:
//
//	<BaseDir>/settings/quickCss.css          Narecord config (installer BaseDir)
//	<BaseDir>/themes/NarecordDen.theme.css
//	<BaseDir>/settings/settings.json         enables QuickCSS + this theme
//
// The shipped desktop.asar is still Equicord-looking and reads Equicord's DATA_DIR
// (EQUICORD_USER_DATA_DIR, else <Discord userData>/../Equicord — typically
// ~/.config/Equicord or %APPDATA%\Equicord), not Narecord. The same files are
// therefore also written to Equicord's config dir so User Settings actually
// shows the den after patch.
const (
	denThemeFileName = "NarecordDen.theme.css"
	denQuickCSSBegin = "/* NARECORD-DEN-BEGIN */"
	denQuickCSSEnd   = "/* NARECORD-DEN-END */"
	denToolboxPlugin = "EquicordToolbox"
)

// Equal-length asar swap. Do not touch EquicordToolbox (plugin id / settings key).
var (
	equicordToolboxLabel = []byte("Equicord Toolbox")
	narecordToolboxLabel = []byte("Narecord Toolbox")
)

//go:embed assets/den/narehate.webp
var denNarehateWebp []byte

//go:embed assets/den/mitty.webp
var denMittyWebp []byte

//go:embed assets/den/narehate-lg.webp
var denNarehateLgWebp []byte

//go:embed assets/den/narehate-circle.webp
var denNarehateCircleWebp []byte

//go:embed assets/den/narehate-circle-lg.webp
var denNarehateCircleLgWebp []byte

//go:embed assets/den/narecord-den.css
var denCSSTemplate string

func dataURI(mime string, b []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b)
}

func denCSS() string {
	// Longer tokens first so __NAREHATE__ cannot eat __NAREHATE_CIRCLE__ / _LG__.
	return strings.NewReplacer(
		"__NAREHATE_CIRCLE_LG__", dataURI("image/webp", denNarehateCircleLgWebp),
		"__NAREHATE_CIRCLE__", dataURI("image/webp", denNarehateCircleWebp),
		"__NAREHATE_LG__", dataURI("image/webp", denNarehateLgWebp),
		"__NAREHATE__", dataURI("image/webp", denNarehateWebp),
		"__MITTY__", dataURI("image/webp", denMittyWebp),
	).Replace(denCSSTemplate)
}

func denThemeCSS() string {
	return `/**
 * @name Narecord Den
 * @author Narehatechi
 * @description Nanachi/Mitty portraits, Narecord Toolbox title-bar, moss/rose hideout chrome, and per-plugin Plugin cards.
 * @version 1.0.0
 */

` + denCSS()
}

func mergeQuickCSS(existing, den string) string {
	block := denQuickCSSBegin + "\n" + den + "\n" + denQuickCSSEnd
	start := strings.Index(existing, denQuickCSSBegin)
	end := strings.Index(existing, denQuickCSSEnd)
	if start >= 0 && end > start {
		return existing[:start] + block + existing[end+len(denQuickCSSEnd):]
	}
	existing = strings.TrimRight(existing, "\n")
	if existing == "" {
		return block + "\n"
	}
	return existing + "\n\n" + block + "\n"
}

func enableDenInSettingsJSON(raw []byte) ([]byte, error) {
	data := map[string]any{}
	if len(strings.TrimSpace(string(raw))) > 0 {
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, fmt.Errorf("parse settings.json: %w", err)
		}
	}
	data["useQuickCss"] = true

	var themes []any
	switch existing := data["enabledThemes"].(type) {
	case []any:
		themes = existing
	case nil:
		themes = nil
	default:
		themes = nil
	}

	found := false
	for _, t := range themes {
		if s, ok := t.(string); ok && s == denThemeFileName {
			found = true
			break
		}
	}
	if !found {
		themes = append(themes, denThemeFileName)
	}
	data["enabledThemes"] = themes

	plugins, _ := data["plugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}
	toolbox, _ := plugins[denToolboxPlugin].(map[string]any)
	if toolbox == nil {
		toolbox = map[string]any{}
	}
	toolbox["enabled"] = true
	plugins[denToolboxPlugin] = toolbox
	data["plugins"] = plugins

	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func denDataDirs() []string {
	seen := map[string]bool{}
	var dirs []string
	add := func(d string) {
		d = strings.TrimSpace(d)
		if d == "" || seen[d] {
			return
		}
		seen[d] = true
		dirs = append(dirs, d)
	}
	add(BaseDir)
	add(appdir.New("Equicord").UserConfig())
	add(os.Getenv("EQUICORD_USER_DATA_DIR"))
	add(os.Getenv("NARECORD_USER_DATA_DIR"))
	return dirs
}

func installDenInto(root string) error {
	if root == "" {
		return errors.New("empty den config dir")
	}
	settingsDir := path.Join(root, "settings")
	themesDir := path.Join(root, "themes")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return err
	}

	css := denCSS()
	themePath := path.Join(themesDir, denThemeFileName)
	if err := os.WriteFile(themePath, []byte(denThemeCSS()), 0644); err != nil {
		return err
	}

	quickPath := path.Join(settingsDir, "quickCss.css")
	existingQuick, err := os.ReadFile(quickPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.WriteFile(quickPath, []byte(mergeQuickCSS(string(existingQuick), css)), 0644); err != nil {
		return err
	}

	settingsPath := path.Join(settingsDir, "settings.json")
	existingSettings, err := os.ReadFile(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	updated, err := enableDenInSettingsJSON(existingSettings)
	if err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, updated, 0644); err != nil {
		return err
	}

	_ = FixOwnership(root)
	_ = FixOwnership(settingsDir)
	_ = FixOwnership(themesDir)
	_ = FixOwnership(themePath)
	_ = FixOwnership(quickPath)
	_ = FixOwnership(settingsPath)
	return nil
}

// InstallDen writes the Narecord settings den (portraits, header, moss/rose chrome,
// banner, Narecord Toolbox, and per-plugin Plugin cards) into Narecord and Equicord
// config dirs so Discord User Settings shows it after this installer patches Discord.
func InstallDen() error {
	css := denCSS()
	if !strings.Contains(css, "Narecord Settings") || !strings.Contains(css, "data:image/webp") {
		return errors.New("den CSS failed to embed portraits")
	}
	if !strings.Contains(css, "Narecord Toolbox") || !strings.Contains(css, "vc-toolbox-btn") {
		return errors.New("den CSS failed to embed Narecord Toolbox")
	}
	if !strings.Contains(css, "Narecord Plugin") || !strings.Contains(css, `alt="User"`) {
		return errors.New("den CSS failed to embed Narecord plugin cards")
	}

	dirs := denDataDirs()
	if len(dirs) == 0 {
		return errors.New("no config dir to install the Narecord settings den into")
	}

	var errs []error
	ok := 0
	for _, dir := range dirs {
		Log.Debug("Installing Narecord settings den into", dir)
		if err := installDenInto(dir); err != nil {
			Log.Warn("Failed to install den into", dir, err)
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		ok++
	}
	if ok == 0 {
		return errors.Join(errs...)
	}
	return nil
}

func patchAsarEquicordToolbox(asarPath string) error {
	if len(equicordToolboxLabel) != len(narecordToolboxLabel) {
		return errors.New("toolbox labels must be the same length to patch asar in place")
	}
	b, err := os.ReadFile(asarPath)
	if err != nil {
		return err
	}
	if !bytes.Contains(b, equicordToolboxLabel) {
		return nil
	}
	patched := bytes.ReplaceAll(b, equicordToolboxLabel, narecordToolboxLabel)
	if bytes.Contains(patched, equicordToolboxLabel) {
		return errors.New("Equicord Toolbox still present after asar patch")
	}
	if err := os.WriteFile(asarPath, patched, 0644); err != nil {
		return err
	}
	_ = FixOwnership(asarPath)
	return nil
}

// PatchShippedAsarToolbox retitles Equicord Toolbox -> Narecord Toolbox inside
// the downloaded desktop.asar. Same length, so asar offsets stay valid.
// Dev directory installs are left alone; QuickCSS still covers the portrait.
func PatchShippedAsarToolbox() error {
	target := NarecordDirectory
	if target == "" {
		return errors.New("empty Narecord directory")
	}
	st, err := os.Stat(target)
	if err != nil {
		return err
	}
	if st.IsDir() {
		Log.Debug("Dev install directory; skipping asar toolbox retitle")
		return nil
	}
	return patchAsarEquicordToolbox(target)
}
