/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import assert from "node:assert/strict";
import { test } from "node:test";

import { checkCompatibility, compareVersions } from "../src/compat.js";
import type { HostBuildIdentity } from "../src/types.js";

const baseIdentity: HostBuildIdentity = {
  platform: "android",
  hostVersion: "1.2.0",
  discordVersion: "250.0",
  bundleFingerprint: "known-good-bundle",
};

test("compareVersions orders dotted versions numerically", () => {
  assert.equal(compareVersions("1.2.0", "1.10.0"), -1);
  assert.equal(compareVersions("1.10.0", "1.2.0"), 1);
  assert.equal(compareVersions("1.2.0", "1.2.0"), 0);
  assert.equal(compareVersions("1.2", "1.2.0"), 0);
});

test("no requirement means compatible", () => {
  assert.deepEqual(checkCompatibility(baseIdentity, undefined), { compatible: true });
});

test("fails closed on platform mismatch", () => {
  const result = checkCompatibility(baseIdentity, { platforms: ["ios"] });
  assert.equal(result.compatible, false);
  assert.match(result.reason ?? "", /platform/);
});

test("fails closed below minimum host version", () => {
  const result = checkCompatibility(baseIdentity, { minHostVersion: "2.0.0" });
  assert.equal(result.compatible, false);
});

test("fails closed above maximum host version", () => {
  const result = checkCompatibility(baseIdentity, { maxHostVersion: "1.0.0" });
  assert.equal(result.compatible, false);
});

test("fails closed on unverified bundle fingerprint", () => {
  const result = checkCompatibility(baseIdentity, {
    allowedBundleFingerprints: ["some-other-bundle"],
  });
  assert.equal(result.compatible, false);
  assert.match(result.reason ?? "", /fingerprint/);
});

test("passes when bundle fingerprint matches allowlist", () => {
  const result = checkCompatibility(baseIdentity, {
    allowedBundleFingerprints: ["known-good-bundle", "another-bundle"],
  });
  assert.equal(result.compatible, true);
});
