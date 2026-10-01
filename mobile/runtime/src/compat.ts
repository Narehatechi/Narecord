/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import type { CompatibilityRequirement, HostBuildIdentity } from "./types.js";

export interface CompatibilityResult {
  compatible: boolean;
  /** Human-readable reason when `compatible` is false. */
  reason?: string;
}

/**
 * Parses a dotted version string ("1.2.3") into an array of numbers,
 * tolerating missing/non-numeric segments by treating them as 0.
 */
function parseVersion(version: string): number[] {
  return version.split(".").map(part => {
    const n = Number.parseInt(part, 10);
    return Number.isFinite(n) ? n : 0;
  });
}

/** Compares two dotted version strings. Returns -1, 0, or 1. */
export function compareVersions(a: string, b: string): number {
  const pa = parseVersion(a);
  const pb = parseVersion(b);
  const len = Math.max(pa.length, pb.length);
  for (let i = 0; i < len; i++) {
    const da = pa[i] ?? 0;
    const db = pb[i] ?? 0;
    if (da !== db) return da < db ? -1 : 1;
  }
  return 0;
}

/**
 * Checks a {@link HostBuildIdentity} against a {@link CompatibilityRequirement}.
 *
 * This is the "fail closed" gate: Discord's mobile bundle layout, anchors,
 * and module ids are version-specific, so if a requirement cannot be
 * positively confirmed (e.g. an unset vs mismatched fingerprint) this
 * function must return `compatible: false` rather than guessing.
 */
export function checkCompatibility(
  identity: HostBuildIdentity,
  requirement: CompatibilityRequirement | undefined,
): CompatibilityResult {
  if (!requirement) {
    return { compatible: true };
  }

  if (requirement.platforms && !requirement.platforms.includes(identity.platform)) {
    return {
      compatible: false,
      reason: `host platform "${identity.platform}" is not in allowed platforms [${requirement.platforms.join(", ")}]`,
    };
  }

  if (
    requirement.minHostVersion &&
    compareVersions(identity.hostVersion, requirement.minHostVersion) < 0
  ) {
    return {
      compatible: false,
      reason: `host version ${identity.hostVersion} is below minimum required ${requirement.minHostVersion}`,
    };
  }

  if (
    requirement.maxHostVersion &&
    compareVersions(identity.hostVersion, requirement.maxHostVersion) > 0
  ) {
    return {
      compatible: false,
      reason: `host version ${identity.hostVersion} is above maximum supported ${requirement.maxHostVersion}`,
    };
  }

  if (
    requirement.allowedBundleFingerprints &&
    !requirement.allowedBundleFingerprints.includes(identity.bundleFingerprint)
  ) {
    return {
      compatible: false,
      reason:
        `Discord bundle fingerprint "${identity.bundleFingerprint}" does not match any ` +
        `verified fingerprint for this plugin/patch. Discord bundle compatibility is ` +
        `version-specific; refusing to load against an unverified bundle.`,
    };
  }

  return { compatible: true };
}
