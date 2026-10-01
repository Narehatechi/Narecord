/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

/**
 * NareBadge — registers a simple "Narecord Mobile" badge through the host
 * UI bridge. Demonstrates the minimal shape of a UI-registering plugin:
 * no patches, no settings, just a badge registration cleaned up on stop.
 */
import type { MobilePlugin, PluginContext, Unregister } from "../../runtime/src/types.js";

let unregisterBadge: Unregister | undefined;

const NareBadge: MobilePlugin = {
  metadata: {
    name: "NareBadge",
    description: "Registers a Narecord Mobile badge in the host UI.",
    authors: [{ name: "Narecord" }],
    version: "0.1.0",
  },
  start(ctx: PluginContext) {
    unregisterBadge = ctx.host.ui.registerBadge({
      id: "narecord-mobile",
      label: "Narecord Mobile",
      icon: "narecord-den-badge",
    });
    ctx.host.log("info", "NareBadge", "badge registered");
  },
  stop(ctx: PluginContext) {
    unregisterBadge?.();
    unregisterBadge = undefined;
    ctx.host.log("info", "NareBadge", "badge unregistered");
  },
};

export default NareBadge;
