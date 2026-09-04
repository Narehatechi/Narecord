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
	"io"
	"os"
	path "path/filepath"
	"strings"
	"sync"

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

var (
	denCSSOnce   sync.Once
	denCSSCached string
)

func denCSS() string {
	denCSSOnce.Do(func() {
		// Longer tokens first so __NAREHATE__ cannot eat __NAREHATE_CIRCLE__ / _LG__.
		denCSSCached = strings.NewReplacer(
			"__NAREHATE_CIRCLE_LG__", dataURI("image/webp", denNarehateCircleLgWebp),
			"__NAREHATE_CIRCLE__", dataURI("image/webp", denNarehateCircleWebp),
			"__NAREHATE_LG__", dataURI("image/webp", denNarehateLgWebp),
			"__NAREHATE__", dataURI("image/webp", denNarehateWebp),
			"__MITTY__", dataURI("image/webp", denMittyWebp),
		).Replace(denCSSTemplate)
	})
	return denCSSCached
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

	// Linux FixOwnership WalkDirs the tree; chowning root covers settings/, themes/,
	// and the files just written. macOS/Windows implementations are no-ops.
	_ = FixOwnership(root)
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

// asarPatchChunkSize is the streaming window for toolbox retitles. Tests shrink it
// to prove matches that span chunk boundaries are still found.
var asarPatchChunkSize = 256 * 1024

func patchAsarEquicordToolbox(asarPath string) error {
	if len(equicordToolboxLabel) != len(narecordToolboxLabel) {
		return errors.New("toolbox labels must be the same length to patch asar in place")
	}
	n, err := replaceAllEqualLengthInPlace(asarPath, equicordToolboxLabel, narecordToolboxLabel)
	if err != nil {
		return err
	}
	if n > 0 {
		_ = FixOwnership(asarPath)
	}
	return nil
}

// replaceAllEqualLengthInPlace overwrites every non-overlapping occurrence of
// needle with repl without loading the file into memory. Lengths must match so
// asar offsets stay valid and a full rewrite is unnecessary.
//
// Replace-all (not replace-first) is required: "Equicord Toolbox" appears more
// than once in desktop.asar (title-bar / tooltip copy). The string does not
// overlap itself, so left-to-right skip-by-len matches bytes.ReplaceAll.
// Absent needle is a read-only no-op success and leaves mode unchanged.
func replaceAllEqualLengthInPlace(name string, needle, repl []byte) (int, error) {
	return replaceAllEqualLengthInPlaceChunked(name, needle, repl, asarPatchChunkSize)
}

func replaceAllEqualLengthInPlaceChunked(name string, needle, repl []byte, chunkSize int) (int, error) {
	if len(needle) != len(repl) {
		return 0, errors.New("needle and replacement must be the same length")
	}
	if len(needle) == 0 {
		return 0, errors.New("empty needle")
	}

	in, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	offs, err := findBytesOffsets(in, needle, chunkSize)
	closeErr := in.Close()
	if err != nil {
		return 0, err
	}
	if closeErr != nil {
		return 0, closeErr
	}
	if len(offs) == 0 {
		return 0, nil
	}

	out, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	for _, off := range offs {
		if _, err := out.WriteAt(repl, off); err != nil {
			return 0, err
		}
	}
	return len(offs), nil
}

func findBytesOffsets(r io.Reader, needle []byte, chunkSize int) ([]int64, error) {
	nlen := len(needle)
	if nlen == 0 {
		return nil, nil
	}
	if chunkSize < nlen {
		chunkSize = nlen
	}
	carry := nlen - 1
	buf := make([]byte, chunkSize)
	window := make([]byte, 0, chunkSize+carry)
	var (
		absStart int64
		offs     []int64
	)
	for {
		n, err := r.Read(buf)
		atEOF := errors.Is(err, io.EOF)
		if n > 0 {
			window = append(window, buf[:n]...)
		}
		if len(window) == 0 {
			if atEOF || (n == 0 && err == nil) {
				break
			}
			if err != nil {
				return nil, err
			}
			continue
		}

		consumed := 0
		for consumed <= len(window)-nlen {
			rel := bytes.Index(window[consumed:], needle)
			if rel < 0 {
				break
			}
			idx := consumed + rel
			offs = append(offs, absStart+int64(idx))
			consumed = idx + nlen
		}

		if atEOF {
			break
		}
		if err != nil {
			return nil, err
		}

		tailStart := len(window) - carry
		if tailStart < consumed {
			tailStart = consumed
		}
		if tailStart < 0 {
			tailStart = 0
		}
		if tailStart > 0 {
			copy(window, window[tailStart:])
			window = window[:len(window)-tailStart]
			absStart += int64(tailStart)
		}
		if n == 0 {
			break
		}
	}
	return offs, nil
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
