/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

/**
 * NareTheme — registers the Narecord "den" color palette as a settings
 * section in the mobile host UI. This mirrors the desktop den theme's
 * Nanachi/Mitty color choices, but only exposes serializable settings
 * data; actual rendering/theming is applied by the host, not by DOM/CSS
 * injection (which does not exist on mobile).
 */
import type { MobilePlugin, PluginContext } from "../../runtime/src/types.js";

export const NARECORD_DEN_COLORS = {
  mossGreen: "#4c6b4f",
  roseAccent: "#c96a7c",
  nanachiBrown: "#7a5230",
  mittyCream: "#e9ddc3",
  abyssIndigo: "#2b2440",
} as const;

export type NareThemeColorKey = keyof typeof NARECORD_DEN_COLORS;

const NareTheme: MobilePlugin = {
  metadata: {
    name: "NareTheme",
    description: "Registers the Narecord den color palette in mobile settings.",
    authors: [{ name: "Narecord" }],
    version: "0.1.0",
  },
  settings: [
    {
      id: "theme",
      title: "Narecord Theme",
      fields: [
        {
          type: "select",
          key: "accentColor",
          label: "Accent color",
          defaultValue: "roseAccent",
          options: Object.entries(NARECORD_DEN_COLORS).map(([key, value]) => ({
            label: `${key} (${value})`,
            value: key,
          })),
        },
        {
          type: "boolean",
          key: "useAbyssBackground",
          label: "Use Abyss indigo background",
          defaultValue: false,
        },
      ],
    },
  ],
  async start(ctx: PluginContext) {
    ctx.host.log("info", "NareTheme", "starting, applying default accent color");
    const stored = await ctx.host.storage.get<string>("NareTheme.accentColor");
    if (!stored) {
      await ctx.host.storage.set("NareTheme.accentColor", "roseAccent" satisfies NareThemeColorKey);
    }
  },
  stop(ctx: PluginContext) {
    ctx.host.log("info", "NareTheme", "stopping");
  },
};

export default NareTheme;
