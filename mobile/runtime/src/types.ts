/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

/**
 * Core type definitions for the portable Narecord mobile plugin runtime.
 *
 * These types intentionally avoid any dependency on Electron, the desktop
 * webpack module graph, or DOM APIs so that they can be shared between an
 * Android host (React Native / Hermes) and a future iOS host.
 */

/** Supported mobile host platforms. "mock" is used by tests and tooling. */
export type HostPlatform = "android" | "ios" | "mock";

/**
 * Identifies the running host build so the runtime can refuse to load
 * plugins (or patches) that were written for an incompatible Discord
 * bundle. Discord's mobile bundle layout and anchors are version-specific,
 * so this identity must be checked *before* any plugin code executes.
 */
export interface HostBuildIdentity {
  /** The host platform the runtime is currently executing on. */
  platform: HostPlatform;
  /** Version of the Narecord mobile host/shell itself (not Discord's). */
  hostVersion: string;
  /** Discord application version string the bundle was extracted from. */
  discordVersion: string;
  /**
   * Opaque identifier (e.g. a hash of known anchor strings) describing the
   * specific Discord bundle this host was built/patched against. Used by
   * the compatibility gate and patch manager to fail closed when the
   * bundle anchors a patch expects are not present.
   */
  bundleFingerprint: string;
}

/** Minimum/maximum compatibility requirements a plugin or patch can declare. */
export interface CompatibilityRequirement {
  /** Platforms this plugin/patch supports. Omit to allow all platforms. */
  platforms?: HostPlatform[];
  /** Minimum host (runtime) version, inclusive, compared as dotted integers. */
  minHostVersion?: string;
  /** Maximum host (runtime) version, inclusive, compared as dotted integers. */
  maxHostVersion?: string;
  /**
   * Exact Discord bundle fingerprints this plugin/patch has been verified
   * against. If present, the bundle fingerprint of the running host MUST
   * be one of these values or loading is refused.
   */
  allowedBundleFingerprints?: string[];
}

export type LogLevel = "debug" | "info" | "warn" | "error";

/** A single notification/toast request surfaced to the user. */
export interface NotificationRequest {
  title: string;
  body: string;
  /** Optional identifier so hosts can de-duplicate/update notifications. */
  id?: string;
}

/** A settings section a plugin wants rendered in the mobile settings UI. */
export interface SettingsSectionRegistration {
  /** Unique id, namespaced by plugin name by the registry. */
  id: string;
  title: string;
  /** Serializable description of settings fields; rendering is host-specific. */
  fields: SettingsFieldDefinition[];
}

export type SettingsFieldDefinition =
  | { type: "boolean"; key: string; label: string; defaultValue: boolean }
  | { type: "string"; key: string; label: string; defaultValue: string }
  | {
      type: "select";
      key: string;
      label: string;
      defaultValue: string;
      options: { label: string; value: string }[];
    };

/** A badge/UI element a plugin wants registered (e.g. a profile badge). */
export interface BadgeRegistration {
  id: string;
  label: string;
  /** Small identifier/name describing the icon/asset to render, host-specific. */
  icon: string;
}

/** Unregister callback returned by UI registration helpers. */
export type Unregister = () => void;

/**
 * Host bridge: the sandboxed surface a plugin is allowed to talk to. This
 * is the only way plugin code touches the outside world - there is no
 * direct DOM, webpack, or native module access.
 */
export interface HostBridge {
  readonly buildIdentity: HostBuildIdentity;

  log(level: LogLevel, scope: string, message: string, ...args: unknown[]): void;

  storage: {
    get<T = unknown>(key: string): Promise<T | undefined>;
    set<T = unknown>(key: string, value: T): Promise<void>;
    delete(key: string): Promise<void>;
  };

  notifications: {
    show(request: NotificationRequest): Promise<void>;
  };

  ui: {
    registerSettingsSection(section: SettingsSectionRegistration): Unregister;
    registerBadge(badge: BadgeRegistration): Unregister;
  };
}

/** A single reversible patch applied through the patch manager. */
export interface PatchDefinition {
  /** Unique id, namespaced by plugin name by the registry. */
  id: string;
  description?: string;
  /** Apply the patch. Must be idempotent-safe to call once per registration. */
  apply(host: HostBridge): void | Promise<void>;
  /** Undo the patch. Called on plugin stop or patch removal. */
  remove(host: HostBridge): void | Promise<void>;
}

/** Abstraction over registering/applying/removing patches. */
export interface PatchManager {
  register(patch: PatchDefinition): void;
  apply(id: string): Promise<void>;
  applyAll(): Promise<void>;
  remove(id: string): Promise<void>;
  removeAll(): Promise<void>;
  has(id: string): boolean;
  list(): string[];
}

/** Context passed into a plugin's lifecycle hooks. */
export interface PluginContext {
  host: HostBridge;
  patches: PatchManager;
}

export interface PluginAuthor {
  name: string;
  id?: string;
}

export interface PluginMetadata {
  name: string;
  description: string;
  authors: PluginAuthor[];
  version?: string;
}

/**
 * A Narecord mobile plugin. Unlike the desktop Vencord/Equicord plugin API,
 * this interface has no `@webpack/common`, no DOM selectors, and no
 * `definePlugin` patch wiring - patches are explicit `PatchDefinition`
 * objects applied through the host-provided `PatchManager`.
 */
export interface MobilePlugin {
  metadata: PluginMetadata;
  /** Names of other plugins (by `metadata.name`) that must start first. */
  dependencies?: string[];
  /** Compatibility requirements gating this plugin from loading at all. */
  compatibility?: CompatibilityRequirement;
  /** Settings sections registered when the plugin starts. */
  settings?: SettingsSectionRegistration[];
  /** Patches registered (but not necessarily applied) when the plugin starts. */
  patches?: PatchDefinition[];
  /** Called once dependencies have started and compatibility has been verified. */
  start(ctx: PluginContext): void | Promise<void>;
  /** Called on shutdown or when the plugin is disabled. Must be safe to call once. */
  stop(ctx: PluginContext): void | Promise<void>;
}
