/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

/**
 * NanachiQuotes — periodically surfaces a Nanachi quote via the host
 * notification bridge. Demonstrates using only `HostBridge` primitives
 * (storage + notifications) with no desktop-only imports (no
 * `@webpack/common`, no DOM, no Electron APIs).
 */
import type { MobilePlugin, PluginContext } from "../../runtime/src/types.js";

const QUOTES = [
  "Nnaa~ don't worry about how it looks.",
  "A hunter's gotta eat, even in the Abyss.",
  "Curses are just the Abyss's way of saying hello.",
  "Whistle rank doesn't matter if your stew is good.",
];

let quoteIndex = 0;

function nextQuote(): string {
  const quote = QUOTES[quoteIndex % QUOTES.length]!;
  quoteIndex += 1;
  return quote;
}

const NanachiQuotes: MobilePlugin = {
  metadata: {
    name: "NanachiQuotes",
    description: "Surfaces a Nanachi quote as a notification on start.",
    authors: [{ name: "Narecord" }],
    version: "0.1.0",
  },
  dependencies: ["NareTheme"],
  async start(ctx: PluginContext) {
    const count = ((await ctx.host.storage.get<number>("NanachiQuotes.count")) ?? 0) + 1;
    await ctx.host.storage.set("NanachiQuotes.count", count);

    const quote = nextQuote();
    ctx.host.log("info", "NanachiQuotes", `start #${count}, quote: ${quote}`);
    await ctx.host.notifications.show({
      id: "nanachi-quote",
      title: "Nanachi says...",
      body: quote,
    });
  },
  stop(ctx: PluginContext) {
    ctx.host.log("info", "NanachiQuotes", "stopping");
  },
};

export default NanachiQuotes;
