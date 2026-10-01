# Narecord Mobile (phase 1 foundation)

This is a **from-scratch mobile Narecord loader**, architecturally distinct
from the existing Go desktop installer (`Narelotl`, the rest of this
repository). It is the first phase of a Revenge-like Discord mobile mod
loader for Narecord, starting Android-first while keeping the runtime and
plugin model portable to a future iOS host.

**This is phase one.** It ships a real, tested, portable plugin runtime
and three example plugins that run entirely in-memory against a mock host.
It does **not** ship a working Android app, Discord bundle patcher, or any
code that touches a real Discord installation. See
["What is / is not implemented"](#what-is--is-not-implemented-in-this-phase)
below for the exact boundary.

## Why this is a separate thing from the desktop installer

| | Desktop (`Narelotl`, Go) | Mobile (`mobile/`, this phase) |
| --- | --- | --- |
| Target | Electron desktop Discord client | Android/iOS Discord app |
| Patch target | `app.asar` (Electron/Node bundle) | React Native/Hermes bundle (not yet patched in this phase) |
| Plugin API | Equicord/Vencord `definePlugin`, `@webpack/common`, DOM | `MobilePlugin` (this runtime): no webpack, no DOM |
| Install mechanism | Go installer patches files on disk | N/A yet — no APK/IPA packaging in this phase |
| Language | Go + embedded TS userplugins | TypeScript runtime/plugins + Kotlin host interface |

The existing `plugins/*.tsx` "den" plugins are written against Equicord's
desktop webpack-find-and-patch API (`@webpack/common`, `Menu`, DOM-based
context menus, etc.). **None of that exists on mobile.** A React
Native/Hermes bundle has no DOM, no desktop webpack module graph, and no
Electron APIs — so mobile plugins are a different interface, not a port.
`mobile/plugins/` are new, from-scratch examples written against the
runtime defined here; they are not copies of `plugins/*.tsx` and not
Vencord-style userplugins.

## Runtime layers

```
mobile/
  runtime/
    src/
      types.ts          Plugin, HostBridge, PatchManager, compatibility types
      hostBridge.ts      InMemoryHostBridge (mock host, used by tests/tooling)
      patchManager.ts    DefaultPatchManager (register/apply/remove patches)
      compat.ts          Host build identity + compatibility gate ("fail closed")
      pluginRegistry.ts  PluginRegistry (load/start/stop plugins)
      index.ts           Barrel export
    test/                node:test unit tests for all of the above
  plugins/
    NareTheme/           settings/theme example plugin
    NanachiQuotes/       notification example plugin (depends on NareTheme)
    NareBadge/           UI badge registration example plugin
  android/
    hostbridge/          Kotlin interface mirror + mock adapter (no app, no SDK build)
```

1. **Host bridge** (`HostBridge`) — the only surface a plugin can touch:
   logging, key/value storage, notifications, and UI registration
   (settings sections, badges). `InMemoryHostBridge` is a full mock
   implementation used by every test in this phase; a real Android host
   bridge (React Native/Hermes-backed) is future work (see roadmap).
2. **Module resolver** — in this phase, "module resolution" is just
   standard Node/TypeScript ESM imports (`mobile/plugins/*/index.ts`
   importing `mobile/runtime/src/types.ts`). There is no custom module
   loader because there is no webpack module graph to resolve against on
   mobile; a real Android host will need its own bundle/module strategy,
   which is explicitly out of scope here (see TODOs under
   `mobile/android/`).
3. **Patch manager** (`PatchManager` / `DefaultPatchManager`) — register,
   apply, and remove named patches. A patch is just an `apply`/`remove`
   function pair that receives the `HostBridge`; nothing assumes a
   desktop webpack/DOM patch target.
4. **Plugin lifecycle** (`PluginRegistry`) — registers plugins, rejects
   duplicates and incompatible hosts before any plugin code runs, starts
   plugins in dependency order, isolates `start`/`stop` failures per
   plugin so one broken plugin cannot take down the others, and always
   cleans up patches/settings registrations on stop (even if `stop()`
   itself throws).
5. **Storage** — `HostBridge.storage` is a minimal async key/value store.
   `InMemoryHostBridge` backs it with a `Map`; a real host would back it
   with platform storage (e.g. `AsyncStorage`/EncryptedSharedPreferences
   on Android).
6. **Settings/UI bridge** — `HostBridge.ui.registerSettingsSection` and
   `registerBadge` let a plugin describe settings fields and badges as
   plain serializable data. Actual rendering is entirely host-specific and
   not implemented here; the mock just records registrations so tests can
   assert on them.

## Compatibility gate ("fail closed")

Discord's mobile bundle layout, module ids, and string anchors are
**version-specific**: a patch or plugin written against one Discord
release can silently corrupt or crash a different release. `compat.ts`
implements `checkCompatibility(hostIdentity, requirement)`, checked by
`PluginRegistry.register()` **before** any plugin code executes. A plugin
can declare:

- `platforms` — restrict to `"android"` / `"ios"` / `"mock"`.
- `minHostVersion` / `maxHostVersion` — restrict to a runtime version range.
- `allowedBundleFingerprints` — restrict to specific, previously-verified
  Discord bundle fingerprints.

If any declared requirement cannot be **positively confirmed**, the
plugin is rejected with `IncompatiblePluginError` — there is no "best
effort" fallback. The real Android host's bundle fingerprint computation
(hashing known anchors from the installed Discord bundle) is **not
implemented in this phase**; see `mobile/android/hostbridge/.../HostBridge.kt`
TODOs.

## What is / is not implemented in this phase

**Implemented and tested:**
- The full TypeScript plugin runtime (`mobile/runtime/`): types, mock host
  bridge, patch manager, plugin registry, compatibility gate.
- Three example plugins (`NareTheme`, `NanachiQuotes`, `NareBadge`) that
  load, start, and stop through the registry in tests.
- A Kotlin interface mirror of the host bridge plus an in-memory mock
  (`mobile/android/hostbridge/`), with no Android SDK/Gradle build
  dependency.
- Node-based unit tests (`node --test`) covering registration, dependency
  ordering, duplicate detection, incompatible-host rejection, patch
  cleanup, start/stop failure isolation, and mock storage.
- A CI job that runs these tests independently of the Go desktop build.

**Explicitly NOT implemented in this phase** (tracked as roadmap items
below, not silently assumed to work):
- Reading, patching, or re-signing any real Discord APK/IPA/bundle.
- Any Hermes/JavaScriptCore runtime embedding or bootstrapping.
- A real Android app module (Gradle/AGP project, `AndroidManifest.xml`,
  UI screens) or any iOS project.
- Per-Discord-version bundle anchor tables or real fingerprint computation.
- APK/IPA packaging, signing, or distribution of any kind.
- Discord account automation, credential handling, or anti-detection
  bypasses of any kind — none of that is in scope for this project ever.
- Any Discord copyrighted assets or bundle contents (none are included
  here, and none should be committed in future phases either).

## Example plugins

- **`NareTheme`** — registers a settings section exposing the Narecord
  "den" color palette (`NARECORD_DEN_COLORS`) as a select field, mirroring
  the desktop den theme's Nanachi/Mitty color choices without any CSS/DOM
  injection.
- **`NanachiQuotes`** — depends on `NareTheme` (demonstrating dependency
  ordering), uses `HostBridge.storage` to count starts and
  `HostBridge.notifications` to show a rotating Nanachi quote. No
  desktop-only imports.
- **`NareBadge`** — the minimal example: registers a single UI badge on
  start and unregisters it on stop.

## Building and testing

All commands below are run from the `mobile/` directory.

```sh
cd mobile
npm install
npm run build   # tsc type-check + compile to dist/
npm test        # node --test against the compiled output (run build first)
```

Requirements: Node.js >= 20 (uses the built-in `node:test` runner — no
Jest/Mocha/ts-node dependency). No Discord installation, Android SDK, or
physical/emulated device is required to build or test anything in this
directory.

## Roadmap (beyond this phase)

1. **Android packaging** — stand up a real Gradle/AGP app module under
   `mobile/android/`, embed Hermes (or use the host app's own Hermes
   instance via JSI), and define a concrete bundle-injection strategy with
   a version-pinned anchor table. Fail closed (refuse to patch) for any
   Discord version without a verified anchor entry.
2. **iOS support** — once the Android host bridge is proven out, implement
   an equivalent Swift/Obj-C host bridge conforming to the same
   `HostBridge` contract, reusing the TypeScript runtime unchanged.
3. **First-party Narecord mobile plugins** — port the *concepts* (not the
   code) of a few desktop den plugins (e.g. a mobile `Abyss` layer tracker)
   to the mobile `MobilePlugin` API once a real host bridge exists to run
   them against.
4. **Signing/distribution** — once bundle injection exists, define a
   signing and installation flow that never automates Discord account
   actions and never redistributes Discord's own assets.
