# Android build

The `app` module is a **buildable Android preview shell**, not a working
Narecord loader or Discord patcher. The launch screen identifies the
unfinished host integration; the Kotlin `HostBridge` remains an interface
and mock. No Discord APK or bundle is read, modified, signed, or distributed.

## Requirements

- JDK 17
- Android SDK with platform 35 and build tools 35.0.0 (the Gradle wrapper
  downloads the pinned Gradle distribution)

## Build

Run these commands from this directory:

```sh
./gradlew :app:assembleDebug
./gradlew :app:assembleRelease
```

The release build is **unsigned** unless all four signing environment
variables below are set. An unsigned APK is suitable for build verification,
not installation or publication.

```sh
export NARECORD_ANDROID_KEYSTORE=/absolute/path/to/release.keystore
export NARECORD_ANDROID_KEYSTORE_PASSWORD='...'
export NARECORD_ANDROID_KEY_ALIAS='...'
export NARECORD_ANDROID_KEY_PASSWORD='...'
./gradlew :app:assembleRelease
```

The corresponding signed APK is written to
`app/build/outputs/apk/release/app-release.apk`; without signing credentials
Gradle writes `app/build/outputs/apk/release/app-release-unsigned.apk`.
Never commit a keystore or its passwords. The Android CI job builds the
unsigned variant for verification only; it does not publish APKs or use
signing secrets. A future distribution job must store the keystore as a
protected secret, decode it to a temporary file, and pass its path and
passwords as environment variables.

Set `-PnarecordVersionCode=<integer>` and
`-PnarecordVersionName=<version>` to override the default app version for a
build. Increase the version code for every published update.

## Not ready for app-store release

This build verifies the Android package and release configuration only. It
does not load the TypeScript runtime, embed Hermes, implement a real
`HostBridge`, inject a bundle, or provide functional loader UI. Those are
separate implementation milestones; do not publish this preview shell as a
working Narecord mobile release.

The Kotlin `HostBridge` types mirror `mobile/runtime/src/types.ts`; the
TypeScript runtime remains the source of truth for that contract.
