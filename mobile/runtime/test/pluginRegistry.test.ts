/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import assert from "node:assert/strict";
import { test } from "node:test";

import { InMemoryHostBridge } from "../src/hostBridge.js";
import {
  CircularDependencyError,
  DuplicatePluginError,
  IncompatiblePluginError,
  PluginRegistry,
  UnknownDependencyError,
} from "../src/pluginRegistry.js";
import type { MobilePlugin, PluginContext } from "../src/types.js";

function makePlugin(overrides: Partial<MobilePlugin> & { name: string }): MobilePlugin {
  const { name, ...rest } = overrides;
  return {
    metadata: { name, description: `${name} test plugin`, authors: [{ name: "test" }] },
    start() {},
    stop() {},
    ...rest,
  };
}

test("registers and starts plugins in dependency order", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  const startOrder: string[] = [];

  registry.register(
    makePlugin({
      name: "Base",
      start: () => {
        startOrder.push("Base");
      },
    }),
  );
  registry.register(
    makePlugin({
      name: "Dependent",
      dependencies: ["Base"],
      start: () => {
        startOrder.push("Dependent");
      },
    }),
  );

  const errors = await registry.startAll();
  assert.deepEqual(errors, []);
  assert.deepEqual(startOrder, ["Base", "Dependent"]);
  assert.equal(registry.isStarted("Base"), true);
  assert.equal(registry.isStarted("Dependent"), true);
});

test("rejects duplicate plugin names", () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  registry.register(makePlugin({ name: "Dup" }));
  assert.throws(() => registry.register(makePlugin({ name: "Dup" })), DuplicatePluginError);
});

test("rejects plugins with unknown dependencies", () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  assert.throws(
    () => registry.register(makePlugin({ name: "Orphan", dependencies: ["Missing"] })),
    UnknownDependencyError,
  );
});

test("detects circular dependencies at start time", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  // Register A depending on B, then B depending on A - the second
  // registration will know "A" exists, forming a cycle only visible
  // once both are registered and we attempt to resolve start order.
  registry.register(makePlugin({ name: "A", dependencies: [] }));
  registry.register(makePlugin({ name: "B", dependencies: ["A"] }));
  // Mutate A's dependencies after the fact to form a cycle (simulating a
  // registry that allowed forward references); this exercises the cycle
  // detector directly via resolveStartOrder through startAll.
  const aEntry = (registry as unknown as { plugins: Map<string, { plugin: MobilePlugin }> }).plugins.get("A");
  if (!aEntry) throw new Error("expected plugin A to be registered");
  aEntry.plugin.dependencies = ["B"];

  await assert.rejects(() => registry.startAll(), CircularDependencyError);
});

test("rejects incompatible host builds before plugin code starts", () => {
  const host = new InMemoryHostBridge({ hostVersion: "1.0.0" });
  const registry = new PluginRegistry(host);
  let started = false;

  assert.throws(() => {
    registry.register(
      makePlugin({
        name: "NeedsNewHost",
        compatibility: { minHostVersion: "2.0.0" },
        start: () => {
          started = true;
        },
      }),
    );
  }, IncompatiblePluginError);

  assert.equal(started, false);
  assert.equal(registry.isRegistered("NeedsNewHost"), false);
});

test("rejects plugins restricted to a different platform", () => {
  const host = new InMemoryHostBridge({ platform: "ios" });
  const registry = new PluginRegistry(host);
  assert.throws(
    () =>
      registry.register(
        makePlugin({ name: "AndroidOnly", compatibility: { platforms: ["android"] } }),
      ),
    IncompatiblePluginError,
  );
});

test("rejects plugins pinned to an unverified bundle fingerprint", () => {
  const host = new InMemoryHostBridge({ bundleFingerprint: "unknown-bundle" });
  const registry = new PluginRegistry(host);
  assert.throws(
    () =>
      registry.register(
        makePlugin({
          name: "PinnedPatch",
          compatibility: { allowedBundleFingerprints: ["known-good-bundle"] },
        }),
      ),
    IncompatiblePluginError,
  );
});

test("isolates start failures so other plugins still start", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  let goodStarted = false;

  registry.register(
    makePlugin({
      name: "Bad",
      start: () => {
        throw new Error("boom");
      },
    }),
  );
  registry.register(
    makePlugin({
      name: "Good",
      start: () => {
        goodStarted = true;
      },
    }),
  );

  const errors = await registry.startAll();
  assert.equal(errors.length, 1);
  assert.equal(errors[0]?.pluginName, "Bad");
  assert.equal(goodStarted, true);
  assert.equal(registry.isStarted("Bad"), false);
  assert.equal(registry.isStarted("Good"), true);
});

test("isolates stop failures and still cleans up patches/settings", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  let patchRemoved = false;

  registry.register(
    makePlugin({
      name: "BadStopper",
      settings: [{ id: "sec", title: "Section", fields: [] }],
      patches: [
        {
          id: "BadStopper.patch",
          apply: () => {},
          remove: () => {
            patchRemoved = true;
          },
        },
      ],
      start: async ctx => {
        await ctx.patches.applyAll();
      },
      stop: () => {
        throw new Error("stop boom");
      },
    }),
  );

  await registry.startAll();
  assert.equal(host.settingsSections.has("sec"), true);

  const stopErrors = await registry.stopAll();
  assert.equal(stopErrors.length, 1);
  assert.equal(stopErrors[0]?.pluginName, "BadStopper");
  assert.equal(patchRemoved, true, "patch should be removed even if stop() throws");
  assert.equal(host.settingsSections.has("sec"), false, "settings section should be unregistered");
  assert.equal(registry.isStarted("BadStopper"), false);
});

test("clean shutdown stops plugins in reverse dependency order", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  const stopOrder: string[] = [];

  registry.register(
    makePlugin({
      name: "Base",
      stop: () => {
        stopOrder.push("Base");
      },
    }),
  );
  registry.register(
    makePlugin({
      name: "Dependent",
      dependencies: ["Base"],
      stop: () => {
        stopOrder.push("Dependent");
      },
    }),
  );

  await registry.startAll();
  const errors = await registry.stopAll();
  assert.deepEqual(errors, []);
  assert.deepEqual(stopOrder, ["Dependent", "Base"]);
});

test("cannot unregister a started plugin without stopping it first", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  registry.register(makePlugin({ name: "Sticky" }));
  await registry.startAll();
  assert.throws(() => registry.unregister("Sticky"));
  await registry.stopAll();
  assert.doesNotThrow(() => registry.unregister("Sticky"));
  assert.equal(registry.isRegistered("Sticky"), false);
});

test("mock storage persists values across get/set/delete", async () => {
  const host = new InMemoryHostBridge();
  assert.equal(await host.storage.get("missing"), undefined);
  await host.storage.set("key", { a: 1 });
  assert.deepEqual(await host.storage.get("key"), { a: 1 });
  await host.storage.delete("key");
  assert.equal(await host.storage.get("key"), undefined);
});

test("mock host bridge records notifications sent by plugins", async () => {
  const host = new InMemoryHostBridge();
  const registry = new PluginRegistry(host);
  registry.register(
    makePlugin({
      name: "Notifier",
      start: async (ctx: PluginContext) => {
        await ctx.host.notifications.show({ title: "Hi", body: "there" });
      },
    }),
  );
  await registry.startAll();
  assert.equal(host.sentNotifications.length, 1);
  assert.equal(host.sentNotifications[0]?.title, "Hi");
});
