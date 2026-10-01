# Android host skeleton (phase 1)

This directory contains a **Kotlin interface skeleton**, not a working
Android application or Discord patcher. It defines the Kotlin-side mirror
of the TypeScript `HostBridge` contract in `../runtime/src/types.ts` so
that future Android integration work has a stable starting point.

## What's here

- `hostbridge/src/main/kotlin/.../HostBridge.kt` — the host bridge
  interface, plus explicit `TODO(...)` markers for everything not yet
  implemented (bundle injection, Hermes integration, APK signing,
  version-specific anchors).
- `hostbridge/src/main/kotlin/.../MockHostBridge.kt` — an in-memory mock
  implementation, mirroring `InMemoryHostBridge` from the TypeScript
  runtime, for future Kotlin-side unit tests.

## What's deliberately NOT here

- No Gradle wrapper, Android Gradle Plugin, or `AndroidManifest.xml`. A
  real Android app module requires the Android SDK/build tools, which are
  not assumed to be available in this repository's CI or sandbox, and
  standing one up prematurely would create an unbuildable, unmaintained
  stub. When Android integration begins in earnest, this directory should
  grow a proper Gradle module (see `mobile/README.md` roadmap).
- No code that reads, patches, re-signs, or redistributes a Discord
  APK/bundle.
- No embedding of a Hermes/JS runtime.

## Relationship to `mobile/runtime`

The Kotlin types in `HostBridge.kt` intentionally mirror the TypeScript
types in `mobile/runtime/src/types.ts` field-for-field. When the two
diverge, the TypeScript runtime is the source of truth (it is the part
that is actually built and tested in CI today); update the Kotlin mirror
to match it.
