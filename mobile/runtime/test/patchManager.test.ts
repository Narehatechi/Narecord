/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import assert from "node:assert/strict";
import { test } from "node:test";

import { InMemoryHostBridge } from "../src/hostBridge.js";
import { DefaultPatchManager } from "../src/patchManager.js";

test("applies and removes a patch exactly once", async () => {
  const host = new InMemoryHostBridge();
  const manager = new DefaultPatchManager(host);
  let applyCount = 0;
  let removeCount = 0;

  manager.register({
    id: "p1",
    apply: () => {
      applyCount++;
    },
    remove: () => {
      removeCount++;
    },
  });

  await manager.apply("p1");
  await manager.apply("p1"); // idempotent, should not double-apply
  assert.equal(applyCount, 1);

  await manager.remove("p1");
  await manager.remove("p1"); // no-op, already removed
  assert.equal(removeCount, 1);
  assert.equal(manager.has("p1"), false);
});

test("rejects duplicate patch ids", () => {
  const host = new InMemoryHostBridge();
  const manager = new DefaultPatchManager(host);
  manager.register({ id: "dup", apply: () => {}, remove: () => {} });
  assert.throws(() => manager.register({ id: "dup", apply: () => {}, remove: () => {} }));
});

test("throws when applying an unregistered patch id", async () => {
  const host = new InMemoryHostBridge();
  const manager = new DefaultPatchManager(host);
  await assert.rejects(() => manager.apply("nope"));
});

test("applyAll/removeAll operate on every registered patch", async () => {
  const host = new InMemoryHostBridge();
  const manager = new DefaultPatchManager(host);
  const applied: string[] = [];
  const removed: string[] = [];

  for (const id of ["a", "b", "c"]) {
    manager.register({
      id,
      apply: () => {
        applied.push(id);
      },
      remove: () => {
        removed.push(id);
      },
    });
  }

  await manager.applyAll();
  assert.deepEqual(applied.sort(), ["a", "b", "c"]);

  await manager.removeAll();
  assert.deepEqual(removed.sort(), ["a", "b", "c"]);
  assert.deepEqual(manager.list(), []);
});
