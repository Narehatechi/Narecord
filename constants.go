package main

import (
"equilotl/buildinfo"
"image/color"
)

const ReleaseUrl = "https://api.github.com/repos/blackperson12121/Narecord/releases/latest"
const ReleaseUrlFallback = "https://api.github.com/repos/blackperson12121/Narecord/releases/latest"
const InstallerReleaseUrl = "https://api.github.com/repos/blackperson12121/Narelotl/releases/latest"
const InstallerReleaseUrlFallback = "https://api.github.com/repos/blackperson12121/Narelotl/releases/latest"

var UserAgent = "Narelotl/" + buildinfo.InstallerGitHash + "[](https://github.com/blackperson12121/Narelotl)"

var (
DiscordGreen  = color.RGBA{R: 0x6C, G: 0x7A, B: 0x56, A: 0xFF}
DiscordRed    = color.RGBA{R: 0x8B, G: 0x3A, B: 0x44, A: 0xFF}
DiscordBlue   = color.RGBA{R: 0x58, G: 0x65, B: 0xF2, A: 0xFF}
DiscordYellow = color.RGBA{R: 0xE4, G: 0xC0, B: 0x4A, A: 0xFF}
)

var LinuxDiscordNames = []string{
"Discord", "DiscordPTB", "DiscordCanary", "DiscordDevelopment",
"discord", "discordptb", "discordcanary", "discorddevelopment",
"discord-ptb", "discord-canary", "discord-development",
"com.discordapp.Discord", "com.discordapp.DiscordPTB",
"com.discordapp.DiscordCanary", "com.discordapp.DiscordDevelopment",
}