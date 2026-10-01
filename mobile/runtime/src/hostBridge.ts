/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

import type {
  BadgeRegistration,
  HostBridge,
  HostBuildIdentity,
  LogLevel,
  NotificationRequest,
  SettingsSectionRegistration,
  Unregister,
} from "./types.js";

export interface LogEntry {
  level: LogLevel;
  scope: string;
  message: string;
  args: unknown[];
}

/**
 * In-memory / mock implementation of {@link HostBridge}. Used by tests and
 * local tooling so the runtime can be exercised without a real Android or
 * iOS host attached. Every mutating call is observable via the exposed
 * arrays/maps so tests can assert on side effects.
 */
export class InMemoryHostBridge implements HostBridge {
  readonly buildIdentity: HostBuildIdentity;
  readonly logs: LogEntry[] = [];
  /** Every notification ever sent through this bridge, in send order. */
  readonly sentNotifications: NotificationRequest[] = [];
  readonly settingsSections = new Map<string, SettingsSectionRegistration>();
  readonly badges = new Map<string, BadgeRegistration>();

  private readonly store = new Map<string, unknown>();

  constructor(buildIdentity: Partial<HostBuildIdentity> = {}) {
    this.buildIdentity = {
      platform: "mock",
      hostVersion: "0.1.0",
      discordVersion: "0.0.0-mock",
      bundleFingerprint: "mock-fingerprint",
      ...buildIdentity,
    };
  }

  log(level: LogLevel, scope: string, message: string, ...args: unknown[]): void {
    this.logs.push({ level, scope, message, args });
  }

  storage = {
    get: async <T = unknown>(key: string): Promise<T | undefined> => {
      return this.store.has(key) ? (this.store.get(key) as T) : undefined;
    },
    set: async <T = unknown>(key: string, value: T): Promise<void> => {
      this.store.set(key, value);
    },
    delete: async (key: string): Promise<void> => {
      this.store.delete(key);
    },
  };

  notifications = {
    show: async (request: NotificationRequest): Promise<void> => {
      this.sentNotifications.push(request);
    },
  };

  ui = {
    registerSettingsSection: (section: SettingsSectionRegistration): Unregister => {
      this.settingsSections.set(section.id, section);
      return () => {
        this.settingsSections.delete(section.id);
      };
    },
    registerBadge: (badge: BadgeRegistration): Unregister => {
      this.badges.set(badge.id, badge);
      return () => {
        this.badges.delete(badge.id);
      };
    },
  };
}
