# Android build guide

English | [简体中文](../../zh/guides/android.md)

On desktop platforms this binding loads the MaaFramework dynamic libraries through [purego](https://github.com/ebitengine/purego) and needs no C compiler. Android is the exception: purego goes through cgo there, so Android builds require `CGO_ENABLED=1` and the clang from the Android NDK as the C compiler.

## Prerequisites

- Go 1.24 or later
- [Android NDK](https://developer.android.com/ndk/downloads); the build uses its clang as `CC`
- MaaFramework Android libraries for the target architecture, available to the program at runtime (see [Runtime libraries](#runtime-libraries))

## Build command

Prefix your usual `go build` with the following environment variables. For an arm64 device:

```shell
GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang CGO_CFLAGS=--target=aarch64-linux-android24 CGO_LDFLAGS=--target=aarch64-linux-android24 go build
```

For an x86_64 emulator, switch `GOARCH` and the target triple to `amd64` / `x86_64`:

```shell
GOOS=android GOARCH=amd64 CGO_ENABLED=1 CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang CGO_CFLAGS=--target=x86_64-linux-android24 CGO_LDFLAGS=--target=x86_64-linux-android24 go build
```

## Command breakdown

| Variable | Purpose |
| --- | --- |
| `GOOS=android GOARCH=arm64` | Target platform and architecture. Use `GOARCH=amd64` for x86_64 emulators. |
| `CGO_ENABLED=1` | Enables cgo, which purego relies on to load libraries on Android. |
| `CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang` | NDK clang as the C compiler. `<ndk>` is the NDK root directory and `<host>` is the subdirectory for the build machine, such as `linux-x86_64`, `darwin-x86_64`, or `windows-x86_64`. |
| `CGO_CFLAGS` / `CGO_LDFLAGS` | Compile and link target triple, `<arch>-linux-android<API level>`. The trailing number is the target Android API level; the examples use 24. |

## Runtime libraries

A program built this way loads four MaaFramework libraries at runtime: `libMaaFramework.so`, `libMaaToolkit.so`, `libMaaAgentServer.so`, and `libMaaAgentClient.so`. All four must come from one compatible MaaFramework release.

Download the archive for the target architecture from the [MaaFramework releases](https://github.com/MaaXYZ/MaaFramework/releases) — `MAA-android-aarch64-*.zip` for arm64 devices, `MAA-android-x86_64-*.zip` for x86_64 emulators — and point the program at the directory containing the four `.so` files:

```go
maa.Init(maa.WithLibDir("path/to/MaaFramework/libs/android"))
```

Without `WithLibDir`, `Init` uses the platform loader's current search configuration, for example `LD_LIBRARY_PATH` or the directories the app process already searches.
