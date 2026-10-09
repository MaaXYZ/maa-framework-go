# Android 构建指南

[English](../../en/guides/android.md) | 简体中文

在桌面平台上，本绑定通过 [purego](https://github.com/ebitengine/purego) 加载 MaaFramework 动态库，不需要 C 编译器。Android 是例外：purego 在 Android 上要借助 cgo，因此构建时需要 `CGO_ENABLED=1`，并使用 Android NDK 中的 clang 作为 C 编译器。

## 前置条件

- Go 1.24 及以上
- [Android NDK](https://developer.android.com/ndk/downloads)，构建使用其中的 clang 作为 `CC`
- 目标架构的 MaaFramework Android 动态库，在运行时可供程序加载（见[运行时库](#运行时库)）

## 构建命令

在常用的 `go build` 前加上以下环境变量。arm64 设备：

```shell
GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang CGO_CFLAGS=--target=aarch64-linux-android24 CGO_LDFLAGS=--target=aarch64-linux-android24 go build
```

x86_64 模拟器把 `GOARCH` 和 target triple 换成 `amd64` / `x86_64`：

```shell
GOOS=android GOARCH=amd64 CGO_ENABLED=1 CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang CGO_CFLAGS=--target=x86_64-linux-android24 CGO_LDFLAGS=--target=x86_64-linux-android24 go build
```

## 命令说明

| 变量 | 作用 |
| --- | --- |
| `GOOS=android GOARCH=arm64` | 目标平台和架构，x86_64 模拟器使用 `GOARCH=amd64`。 |
| `CGO_ENABLED=1` | 启用 cgo；purego 在 Android 上依靠 cgo 加载动态库。 |
| `CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang` | 使用 NDK 的 clang 作为 C 编译器。`<ndk>` 是 NDK 根目录，`<host>` 是构建机对应的子目录，如 `linux-x86_64`、`darwin-x86_64`、`windows-x86_64`。 |
| `CGO_CFLAGS` / `CGO_LDFLAGS` | 编译与链接的 target triple，形如 `<arch>-linux-android<API level>`；末尾数字是目标 Android API level，示例使用 24。 |

## 运行时库

按上述方式构建的程序在运行时加载四个 MaaFramework 动态库：`libMaaFramework.so`、`libMaaToolkit.so`、`libMaaAgentServer.so` 和 `libMaaAgentClient.so`。四个库必须来自同一个兼容的 MaaFramework 版本。

从 [MaaFramework releases](https://github.com/MaaXYZ/MaaFramework/releases) 下载目标架构的压缩包（arm64 设备用 `MAA-android-aarch64-*.zip`，x86_64 模拟器用 `MAA-android-x86_64-*.zip`），并通过 `WithLibDir` 把程序指向包含这四个 `.so` 文件的目录：

```go
if err := maa.Init(maa.WithLibDir("path/to/MaaFramework/libs/android")); err != nil {
	// 根据需要处理该错误，例如检查库目录、设备 ABI 和动态库依赖。
}
```

不传 `WithLibDir` 时，`Init` 使用平台加载器当前的查找配置，例如 `LD_LIBRARY_PATH` 或应用进程默认搜索的目录。
