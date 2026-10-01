plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

val signingEnvironment = listOf(
    "NARECORD_ANDROID_KEYSTORE",
    "NARECORD_ANDROID_KEYSTORE_PASSWORD",
    "NARECORD_ANDROID_KEY_ALIAS",
    "NARECORD_ANDROID_KEY_PASSWORD",
).associateWith { providers.environmentVariable(it).orNull }
val configuredSigningValues = signingEnvironment.values.count { !it.isNullOrBlank() }
require(configuredSigningValues == 0 || configuredSigningValues == signingEnvironment.size) {
    "Set all four NARECORD_ANDROID_KEYSTORE and NARECORD_ANDROID_KEY_* variables to sign a release."
}
val releaseSigningConfigured = configuredSigningValues == signingEnvironment.size

android {
    namespace = "dev.narehatechi.narecord.mobile"
    compileSdk = 35
    buildToolsVersion = "35.0.0"

    defaultConfig {
        applicationId = "dev.narehatechi.narecord.mobile"
        minSdk = 23
        targetSdk = 35
        versionCode = providers.gradleProperty("narecordVersionCode").orElse("1").get().toInt()
        versionName = providers.gradleProperty("narecordVersionName").orElse("0.1.0").get()
    }

    sourceSets["main"].java.srcDir("../hostbridge/src/main/kotlin")

    signingConfigs {
        if (releaseSigningConfigured) {
            create("release") {
                storeFile = file(signingEnvironment.getValue("NARECORD_ANDROID_KEYSTORE")!!)
                storePassword = signingEnvironment.getValue("NARECORD_ANDROID_KEYSTORE_PASSWORD")
                keyAlias = signingEnvironment.getValue("NARECORD_ANDROID_KEY_ALIAS")
                keyPassword = signingEnvironment.getValue("NARECORD_ANDROID_KEY_PASSWORD")
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            signingConfig = if (releaseSigningConfigured) {
                signingConfigs.getByName("release")
            } else {
                null
            }
        }
    }
}
