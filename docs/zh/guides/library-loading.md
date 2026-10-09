# 运行时库加载

[English](../../en/guides/library-loading.md) | 简体中文

Go 包在运行时加载 MaaFramework 动态库，`go get` 不会安装这些库。从 [MaaFramework releases](https://github.com/MaaXYZ/MaaFramework/releases) 下载对应操作系统和架构的压缩包；平台选择见[安装表格](../../../README_zh.md#2-下载-maaframework)，版本选择见[版本策略](../../../README_zh.md#要求)。

## 库文件

`Init` 需要全部四个库，即使程序不使用 Agent：

| 库 | Windows | Linux / Android | macOS |
| --- | --- | --- | --- |
| MaaFramework | `MaaFramework.dll` | `libMaaFramework.so` | `libMaaFramework.dylib` |
| MaaToolkit | `MaaToolkit.dll` | `libMaaToolkit.so` | `libMaaToolkit.dylib` |
| MaaAgentServer | `MaaAgentServer.dll` | `libMaaAgentServer.so` | `libMaaAgentServer.dylib` |
| MaaAgentClient | `MaaAgentClient.dll` | `libMaaAgentClient.so` | `libMaaAgentClient.dylib` |

四个库必须来自同一个兼容的 MaaFramework 版本。完整解压发布包，保留其依赖和目录结构；只复制这四个文件可能导致其原生依赖缺失。兼容性契约见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init)。

桌面平台发布包的运行时库位于 `bin/`。Android 的库目录和构建方式见 [Android 构建指南](android.md)。

## 选择加载方式

根据部署环境，从以下方式中选择一种。

### 指定库目录

传入包含库文件的目录，而不是发布包根目录或单个库文件名：

```go
if err := maa.Init(maa.WithLibDir("path/to/MaaFramework/bin")); err != nil {
	return err
}
```

绝对路径可避免对启动目录的依赖。相对路径以进程的工作目录为基准解析。在较大的应用中嵌入本绑定时，请查阅 [WithLibDir](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#WithLibDir) 中 Windows DLL 搜索配置对整个进程的影响。

### 使用工作目录

将运行时库及其依赖放入启动程序时的工作目录，并显式选择该目录：

```go
if err := maa.Init(maa.WithLibDir(".")); err != nil {
	return err
}
```

工作目录可能与可执行文件目录不同，尤其是在 IDE 或服务中启动时。不传 `WithLibDir` 的 `Init()` 使用平台加载器当前的查找配置，不能保证所有平台都会搜索工作目录。

### 配置环境变量

在启动程序前设置搜索路径，然后调用不带库目录选项的 `maa.Init()`。请使用解压后的库目录的绝对路径。

Windows（PowerShell）：

```powershell
$env:PATH = "C:\MaaFramework\bin;" + $env:PATH
.\your-program.exe
```

Linux（shell）：

```shell
LD_LIBRARY_PATH="/absolute/path/to/MaaFramework/bin${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}" ./your-program
```

macOS（shell）：

```shell
DYLD_LIBRARY_PATH="/absolute/path/to/MaaFramework/bin${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}" ./your-program
```

这些路径遵循操作系统加载器的规则：[Windows DLL 搜索顺序](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-search-order)、[Linux dlopen](https://man7.org/linux/man-pages/man3/dlopen.3.html) 和 [macOS 动态库加载](https://developer.apple.com/library/archive/documentation/DeveloperTools/Conceptual/DynamicLibraries/100-Articles/DynamicLibraryUsageGuidelines.html)。受保护的 macOS 进程可能会移除 `DYLD_*` 变量，见 Apple 的[运行时保护说明](https://developer.apple.com/library/archive/documentation/Security/Conceptual/System_Integrity_Protection_Guide/RuntimeProtections/RuntimeProtections.html)。如果启动器不保留该变量，改用 `WithLibDir` 指定目录。

### 使用系统库目录

集中部署时，可将库及其依赖安装到平台加载器搜索的目录，并调用不带库目录选项的 `maa.Init()`。使用操作系统的部署工具配置这些目录，具体规则见上方加载器文档。升级时检查搜索路径中是否仍有旧版 MaaFramework。

## 排查问题

初始化报告缺库、缺符号或清理错误时，按 [FAQ](faq.md) 排查。初始化和释放的契约见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) 与 [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release)。
