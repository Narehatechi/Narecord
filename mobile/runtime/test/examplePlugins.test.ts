/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import assert from "node:assert/strict";
import { test } from "node:test";

import { InMemoryHostBridge } from "../src/hostBridge.js";
import { PluginRegistry } from "../src/pluginRegistry.js";
import NanachiQuotes from "../../plugins/NanachiQuotes/index.js";
import NareBadge from "../../plugins/NareBadge/index.js";
import NareTheme from "../../plugins/NareTheme/index.js";

test("all three example plugins load, start, and stop through the registry", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);

  registry.register(NareTheme);
  registry.register(NanachiQuotes);
  registry.register(NareBadge);

  const startErrors = await registry.startAll();
  assert.deepEqual(startErrors, []);

  assert.equal(registry.isStarted("NareTheme"), true);
  assert.equal(registry.isStarted("NanachiQuotes"), true);
  assert.equal(registry.isStarted("NareBadge"), true);

  // NanachiQuotes depends on NareTheme and should have sent a notification.
  assert.equal(host.sentNotifications.length, 1);
  assert.equal(host.sentNotifications[0]?.title, "Nanachi says...");

  // NareBadge should have registered a badge.
  assert.equal(host.badges.has("narecord-mobile"), true);

  const stopErrors = await registry.stopAll();
  assert.deepEqual(stopErrors, []);
  assert.equal(host.badges.has("narecord-mobile"), false);
});
