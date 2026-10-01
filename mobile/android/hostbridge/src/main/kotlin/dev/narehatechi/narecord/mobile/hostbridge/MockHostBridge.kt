/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */
package dev.narehatechi.narecord.mobile.hostbridge

/**
 * In-memory mock implementation of [HostBridge], mirroring
 * `InMemoryHostBridge` in `mobile/runtime/src/hostBridge.ts`.
 *
 * This is NOT a real Android host integration. It exists so that future
 * Kotlin-side code (settings screens, notification plumbing, etc.) can be
 * developed and unit-tested against a stable interface before the real
 * Hermes/bundle integration (see TODOs in [HostBridge]) exists.
 */
class MockHostBridge(
    override val buildIdentity: HostBuildIdentity = HostBuildIdentity(
        platform = HostPlatform.MOCK,
        hostVersion = "0.1.0",
        discordVersion = "0.0.0-mock",
        bundleFingerprint = "mock-fingerprint",
    ),
) : HostBridge {
    val logs = mutableListOf<Triple<LogLevel, String, String>>()
    val sentNotifications = mutableListOf<NotificationRequest>()
    val settingsSections = mutableMapOf<String, String>()
    val badges = mutableMapOf<String, Pair<String, String>>()

    private val store = mutableMapOf<String, String>()

    override fun log(level: LogLevel, scope: String, message: String) {
        logs.add(Triple(level, scope, message))
    }

    override suspend fun storageGet(key: String): String? = store[key]

    override suspend fun storageSet(key: String, value: String) {
        store[key] = value
    }

    override suspend fun storageDelete(key: String) {
        store.remove(key)
    }

    override suspend fun showNotification(request: NotificationRequest) {
        sentNotifications.add(request)
    }

    override fun registerSettingsSectionId(id: String, title: String): () -> Unit {
        settingsSections[id] = title
        return { settingsSections.remove(id) }
    }

    override fun registerBadgeId(id: String, label: String, icon: String): () -> Unit {
        badges[id] = label to icon
        return { badges.remove(id) }
    }
}
