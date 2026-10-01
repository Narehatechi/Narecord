/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import { checkCompatibility } from "./compat.js";
import { DefaultPatchManager } from "./patchManager.js";
import type { HostBridge, MobilePlugin, PluginContext, Unregister } from "./types.js";

export class DuplicatePluginError extends Error {
  constructor(name: string) {
    super(`Plugin "${name}" is already registered`);
    this.name = "DuplicatePluginError";
  }
}

export class IncompatiblePluginError extends Error {
  constructor(
    public readonly pluginName: string,
    public readonly reason: string,
  ) {
    super(`Plugin "${pluginName}" is incompatible with this host: ${reason}`);
    this.name = "IncompatiblePluginError";
  }
}

export class UnknownDependencyError extends Error {
  constructor(pluginName: string, dependency: string) {
    super(`Plugin "${pluginName}" depends on unknown plugin "${dependency}"`);
    this.name = "UnknownDependencyError";
  }
}

export class CircularDependencyError extends Error {
  constructor(cycle: string[]) {
    super(`Circular plugin dependency detected: ${cycle.join(" -> ")}`);
    this.name = "CircularDependencyError";
  }
}

/** Per-plugin outcome of a `startAll`/`stopAll` sweep. */
export interface PluginLifecycleError {
  pluginName: string;
  error: unknown;
}

interface RegisteredPlugin {
  plugin: MobilePlugin;
  context: PluginContext;
  settingsUnregisters: Unregister[];
  started: boolean;
}

/**
 * Pure topological sort used to compute plugin start/stop order. Exported
 * standalone (rather than as a private method) so cycle/ordering behavior
 * can be unit-tested directly, without reaching into `PluginRegistry`
 * internals.
 *
 * @param names - all node names to include in the sort.
 * @param getDependencies - returns the declared dependencies of a node.
 * @throws {CircularDependencyError} if the dependency graph has a cycle.
 */
export function topologicalSort(
  names: string[],
  getDependencies: (name: string) => string[],
): string[] {
  const order: string[] = [];
  const visited = new Set<string>();
  const visiting = new Set<string>();
  const path: string[] = [];

  const visit = (name: string): void => {
    if (visited.has(name)) return;
    if (visiting.has(name)) {
      throw new CircularDependencyError([...path, name]);
    }
    visiting.add(name);
    path.push(name);
    for (const dependency of getDependencies(name)) {
      visit(dependency);
    }
    path.pop();
    visiting.delete(name);
    visited.add(name);
    order.push(name);
  };

  for (const name of names) {
    visit(name);
  }

  return order;
}


/**
 * Plugin registry/loader for the mobile runtime.
 *
 * Responsibilities:
 *  - reject duplicate plugin names
 *  - reject plugins incompatible with the current {@link HostBuildIdentity}
 *    before any plugin code runs ("fail closed")
 *  - start plugins in dependency order (topological sort), detecting
 *    unknown dependencies and cycles
 *  - isolate lifecycle errors so one plugin failing to start/stop does not
 *    prevent the others from starting/stopping
 *  - clean up registered patches and settings sections on stop
 */
export class PluginRegistry {
  private readonly plugins = new Map<string, RegisteredPlugin>();

  constructor(private readonly host: HostBridge) {}

  /** Names of all currently registered plugins, in registration order. */
  names(): string[] {
    return [...this.plugins.keys()];
  }

  isRegistered(name: string): boolean {
    return this.plugins.has(name);
  }

  isStarted(name: string): boolean {
    return this.plugins.get(name)?.started ?? false;
  }

  /**
   * Registers a plugin. Throws synchronously (before any plugin code runs)
   * if the name is a duplicate, a dependency is unknown, or the plugin is
   * incompatible with the current host build.
   */
  register(plugin: MobilePlugin): void {
    const name = plugin.metadata.name;
    if (this.plugins.has(name)) {
      throw new DuplicatePluginError(name);
    }

    const compat = checkCompatibility(this.host.buildIdentity, plugin.compatibility);
    if (!compat.compatible) {
      throw new IncompatiblePluginError(name, compat.reason ?? "unknown reason");
    }

    for (const dependency of plugin.dependencies ?? []) {
      if (!this.plugins.has(dependency)) {
        throw new UnknownDependencyError(name, dependency);
      }
    }

    const patches = new DefaultPatchManager(this.host);
    for (const patch of plugin.patches ?? []) {
      patches.register(patch);
    }

    this.plugins.set(name, {
      plugin,
      context: { host: this.host, patches },
      settingsUnregisters: [],
      started: false,
    });
  }

  /** Computes a dependency-respecting start order for all registered plugins. */
  private resolveStartOrder(): string[] {
    return topologicalSort(
      [...this.plugins.keys()],
      name => this.plugins.get(name)?.plugin.dependencies ?? [],
    );
  }

  /**
   * Starts every registered plugin in dependency order. Each plugin's
   * `start` failure is isolated: it is recorded and the remaining plugins
   * still attempt to start. Returns the list of failures (empty on full
   * success).
   */
  async startAll(): Promise<PluginLifecycleError[]> {
    const order = this.resolveStartOrder();
    const errors: PluginLifecycleError[] = [];

    for (const name of order) {
      const entry = this.plugins.get(name);
      if (!entry || entry.started) continue;

      try {
        for (const section of entry.plugin.settings ?? []) {
          entry.settingsUnregisters.push(this.host.ui.registerSettingsSection(section));
        }
        await entry.plugin.start(entry.context);
        entry.started = true;
      } catch (error) {
        errors.push({ pluginName: name, error });
        // Roll back any settings sections registered before the failure so
        // a half-started plugin does not leave stray UI behind.
        for (const unregister of entry.settingsUnregisters.splice(0)) {
          unregister();
        }
      }
    }

    return errors;
  }

  /**
   * Stops every started plugin in reverse dependency order, removing its
   * patches and settings sections regardless of whether `stop` itself
   * throws. Returns the list of `stop` failures (empty on full success).
   */
  async stopAll(): Promise<PluginLifecycleError[]> {
    const order = this.resolveStartOrder().reverse();
    const errors: PluginLifecycleError[] = [];

    for (const name of order) {
      const entry = this.plugins.get(name);
      if (!entry || !entry.started) continue;

      try {
        await entry.plugin.stop(entry.context);
      } catch (error) {
        errors.push({ pluginName: name, error });
      } finally {
        await entry.context.patches.removeAll();
        for (const unregister of entry.settingsUnregisters.splice(0)) {
          unregister();
        }
        entry.started = false;
      }
    }

    return errors;
  }

  /** Removes a plugin's registration entirely. The plugin must be stopped first. */
  unregister(name: string): void {
    const entry = this.plugins.get(name);
    if (!entry) return;
    if (entry.started) {
      throw new Error(`Cannot unregister started plugin "${name}"; stop it first`);
    }
    this.plugins.delete(name);
  }
}
