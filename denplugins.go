/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"fmt"
	"io/fs"
	"os"
	path "path/filepath"
	"strings"

	"embed"
)

// Den userplugin sources live under plugins/<Name>/ and are written on Install
// for source-tree builds. Settings enable is MERGE-ONLY via enableNamedPlugins.
//
// FIXME(stock-asar): Equicord's released desktop.asar does NOT compile or load
// TypeScript from %APPDATA%\Narecord\userplugins (or Equicord\userplugins).
// Live Plugin Management shows Total Userplugins: 0 with sources on disk.
// narePerf only "works" on stock Install via its QuickCSS fallback, not as a
// real userplugin. These plugins appear in Plugin Management only after they
// are compiled into a custom narecord.asar / desktop.asar (userPlugin:true),
// typically by copying this folder into Equicord src/userplugins and running
// `pnpm build`. A follow-up should ship that asar from Narelotl instead of
// stock Equicord desktop.asar.
//
//go:embed plugins/Abyss plugins/Hideout plugins/Incinerator plugins/NareMotion plugins/NarehateBadge plugins/Narelogs plugins/NareNotes plugins/Nnaa
var denUserpluginFS embed.FS

var denUserpluginNames = []string{
	"Abyss",
	"Hideout",
	"Incinerator",
	"NareMotion",
	"NarehateBadge",
	"Narelogs",
	"NareNotes",
	"Nnaa",
}

func isDenUserpluginName(name string) bool {
	for _, n := range denUserpluginNames {
		if n == name {
			return true
		}
	}
	return false
}

func denPluginEmbedDir(name string) string {
	return "plugins/" + name
}

func denPluginIndexSource(name string) (string, error) {
	for _, file := range []string{"index.tsx", "index.ts"} {
		b, err := denUserpluginFS.ReadFile(denPluginEmbedDir(name) + "/" + file)
		if err == nil {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("%s: missing embedded index.ts/tsx", name)
}

func denUserpluginsValid() error {
	for _, name := range denUserpluginNames {
		src, err := denPluginIndexSource(name)
		if err != nil {
			return err
		}
		if !strings.Contains(src, "definePlugin") {
			return fmt.Errorf("%s plugin source missing definePlugin", name)
		}
		if !strings.Contains(src, `name: "`+name+`"`) {
			return fmt.Errorf("%s plugin source missing name: %q", name, name)
		}
	}
	return nil
}

func writeDenUserpluginFiles(dir, name string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	entries, err := fs.ReadDir(denUserpluginFS, denPluginEmbedDir(name))
	if err != nil {
		return err
	}
	wrote := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		rel := denPluginEmbedDir(name) + "/" + e.Name()
		b, err := denUserpluginFS.ReadFile(rel)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(dir, e.Name()), b, 0644); err != nil {
			return err
		}
		wrote++
	}
	if wrote == 0 {
		return fmt.Errorf("%s: no embedded files to write", name)
	}
	return nil
}

func writeDenUserplugins(root string) error {
	for _, name := range denUserpluginNames {
		dir := path.Join(root, "userplugins", name)
		if err := writeDenUserpluginFiles(dir, name); err != nil {
			return err
		}
	}
	return nil
}

func writeDenUserpluginsIntoSourceTree(root string) error {
	if root == "" || !looksLikeEquicordSource(root) {
		return nil
	}
	for _, name := range denUserpluginNames {
		dir := path.Join(root, "src", "userplugins", name)
		if err := writeDenUserpluginFiles(dir, name); err != nil {
			return err
		}
	}
	return nil
}

func installDenUserpluginSources() {
	for _, root := range narePerfSourceRoots() {
		if err := writeDenUserpluginsIntoSourceTree(root); err != nil {
			Log.Warn("Could not write den userplugins into Equicord source", root, err)
		} else if looksLikeEquicordSource(root) {
			Log.Info("Wrote den userplugins into Equicord source", root)
		}
	}
}

func enableDenUserpluginsInSettings(data map[string]any) error {
	return enableNamedPlugins(data, denUserpluginNames...)
}
