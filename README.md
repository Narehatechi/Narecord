# Narelotl

Cross-platform GUI/CLI installer for [Narecord](https://github.com/Narehatechi/Narecord), Narehatechi's Nanachi-themed Equicord/Vencord fork.

This is a rebrand of [Equilotl](https://github.com/Equicord/Equilotl) (itself a fork of the Vencord installer). GPL-3.0.

After Install, Discord User Settings should show **Narecord Settings** with a Nanachi/Narehate portrait on the first row (not an Equicord gear), Mitty on Plugins, moss/rose selected-row chrome, and a den banner on the Narecord page. On the Plugins tab, Narecord's set is labeled **Narecord Plugin** (not Equicord's User Plugin mark), and each of those cards has its own look (Abyss is the layers, NareNotes is a notebook, and so on). Real Equicord/Vencord plugin cards stay as they are.

## Usage

Windows

- [GUI](https://github.com/Narehatechi/Narecord/releases/latest/download/Narelotl.exe)
- [CLI](https://github.com/Narehatechi/Narecord/releases/latest/download/NarelotlCli.exe)

Windows may show "Windows protected your PC" / SmartScreen because Narelotl.exe is unsigned. That warning is harmless — click More info, then Run anyway. Signing would cost money; we are not signing Windows builds.

MacOS

- [X64 GUI](https://github.com/Narehatechi/Narecord/releases/latest/download/Narelotl-darwin-x64.zip)
- [ARM64 GUI](https://github.com/Narehatechi/Narecord/releases/latest/download/Narelotl-darwin-arm64.zip)

Linux

- [GUI](https://github.com/Narehatechi/Narecord/releases/latest/download/Narelotl-x11)
- [CLI](https://github.com/Narehatechi/Narecord/releases/latest/download/NarelotlCli-Linux)

## Building from source

### Prerequisites

You need to install the [Go programming language](https://go.dev/doc/install) and GCC, the GNU Compiler Collection (MinGW on Windows)

<details>
<summary>Additionally, if you're using Linux, you have to install some additional dependencies:</summary>

#### Base dependencies

```sh
apt install -y pkg-config libsdl2-dev libglx-dev libgl1-mesa-dev
dnf install pkg-config libGL-devel libXxf86vm-devel
```

#### X11 dependencies

```sh
apt install -y xorg-dev
dnf install libXcursor-devel libXi-devel libXinerama-devel libXrandr-devel
```

#### Wayland dependencies

```sh
apt install -y libwayland-dev libxkbcommon-dev wayland-protocols extra-cmake-modules
dnf install wayland-devel libxkbcommon-devel wayland-protocols-devel extra-cmake-modules
```

</details>

### Building

#### Install dependencies

```sh
go mod tidy
```

#### Build the GUI

##### Windows / Mac / Linux X11

```sh
go build
```

##### Linux Wayland

```sh
go build --tags wayland
```

#### Build the CLI

```
go build --tags cli
```

You might want to pass some flags to this command to get a better build.
See [the GitHub workflow](https://github.com/Narehatechi/Narecord/blob/main/.github/workflows/release.yml) for what flags I pass or if you want more precise instructions
