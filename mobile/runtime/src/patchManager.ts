/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import type { HostBridge, PatchDefinition, PatchManager } from "./types.js";

/**
 * Default {@link PatchManager} implementation. Patches are tracked by id so
 * they can be applied and removed independently of plugin lifecycle,
 * without assuming any desktop webpack/DOM patch mechanism - applying a
 * patch is just calling the plugin-supplied `apply`/`remove` functions
 * against the host bridge.
 */
export class DefaultPatchManager implements PatchManager {
  private readonly patches = new Map<string, PatchDefinition>();
  private readonly applied = new Set<string>();

  constructor(private readonly host: HostBridge) {}

  register(patch: PatchDefinition): void {
    if (this.patches.has(patch.id)) {
      throw new Error(`Patch with id "${patch.id}" is already registered`);
    }
    this.patches.set(patch.id, patch);
  }

  async apply(id: string): Promise<void> {
    const patch = this.requirePatch(id);
    if (this.applied.has(id)) return;
    await patch.apply(this.host);
    this.applied.add(id);
  }

  async applyAll(): Promise<void> {
    for (const id of this.patches.keys()) {
      await this.apply(id);
    }
  }

  async remove(id: string): Promise<void> {
    const patch = this.patches.get(id);
    if (!patch) return;
    if (this.applied.has(id)) {
      await patch.remove(this.host);
      this.applied.delete(id);
    }
    this.patches.delete(id);
  }

  async removeAll(): Promise<void> {
    for (const id of [...this.patches.keys()]) {
      await this.remove(id);
    }
  }

  has(id: string): boolean {
    return this.patches.has(id);
  }

  list(): string[] {
    return [...this.patches.keys()];
  }

  private requirePatch(id: string): PatchDefinition {
    const patch = this.patches.get(id);
    if (!patch) {
      throw new Error(`No patch registered with id "${id}"`);
    }
    return patch;
  }
}
