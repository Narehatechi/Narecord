/*
 * Narecord Mobile, a from-scratch mobile Narecord loader
 * Copyright (c) 2026 Narehatechi and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */
package dev.narehatechi.narecord.mobile.hostbridge

/**
 * Supported mobile host platforms. Mirrors `HostPlatform` in
 * `mobile/runtime/src/types.ts` - keep these in sync.
 */
enum class HostPlatform {
    ANDROID,
    IOS,
    MOCK,
}

/**
 * Identifies the running host build, including the Discord bundle it was
 * built/verified against. Mirrors `HostBuildIdentity` in
 * `mobile/runtime/src/types.ts`.
 *
 * TODO(android-bundle-fingerprint): [email protected] does not yet compute a real
 * fingerprint from the installed Discord APK/bundle. A real implementation
 * must hash known anchor strings/bytecode offsets from the patched bundle
 * and MUST refuse to report a fingerprint it cannot verify - never fall
 * back to a "best guess" value here.
 */
data class HostBuildIdentity(
    val platform: HostPlatform,
    val hostVersion: String,
    val discordVersion: String,
    val bundleFingerprint: String,
)

enum class LogLevel { DEBUG, INFO, WARN, ERROR }

data class NotificationRequest(
    val title: String,
    val body: String,
    val id: String? = null,
)

/**
 * Kotlin-side mirror of the TypeScript `HostBridge` interface. This is the
 * *only* surface a plugin/patch is allowed to call into on Android.
 *
 * This interface intentionally has NO implementation here that touches a
 * real Discord process, Hermes runtime, or APK contents. See the TODOs
 * below and in `mobile/README.md` for the (currently unimplemented) work
 * required to wire this up to a real host.
 */
interface HostBridge {
    val buildIdentity: HostBuildIdentity

    fun log(level: LogLevel, scope: String, message: String)

    suspend fun storageGet(key: String): String?
    suspend fun storageSet(key: String, value: String)
    suspend fun storageDelete(key: String)

    suspend fun showNotification(request: NotificationRequest)

    // TODO(android-ui-bridge): real settings/badge UI registration requires
    // a live connection to the host's React Native component tree. Phase 1
    // only defines the shape; no Compose/RN view is registered yet.
    fun registerSettingsSectionId(id: String, title: String): () -> Unit
    fun registerBadgeId(id: String, label: String, icon: String): () -> Unit
}

// ---------------------------------------------------------------------------
// Explicit phase-1 boundaries. These TODOs are intentionally NOT implemented
// in this change; see mobile/README.md "What is / is not implemented".
// ---------------------------------------------------------------------------
//
// TODO(bundle-injection): No code here reads, patches, or repackages a
// Discord APK/bundle. Doing so requires a version-pinned map of bundle
// anchors (module ids, known strings) per supported Discord release, and
// must fail closed (refuse to patch) when anchors are not found verbatim.
//
// TODO(hermes-runtime-integration): No code here embeds or starts a Hermes
// JS engine, nor loads the `mobile/runtime` TypeScript build into one. A
// real host needs a JSI/Hermes bridge translating this Kotlin interface
// into the JS `HostBridge` shape consumed by `mobile/runtime/src/types.ts`.
//
// TODO(apk-signing): No code here re-signs or redistributes an APK. Any
// future installer flow must support user-provided signing keys and must
// not embed, redistribute, or automate retrieval of Discord's own assets
// or signing material.
//
// TODO(version-anchors): No per-Discord-version anchor table exists yet.
// `HostBuildIdentity.bundleFingerprint` must be computed from such a table
// before this bridge can be considered safe to apply real patches with.
