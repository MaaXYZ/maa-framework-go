# Runtime library loading

English | [简体中文](../../zh/guides/library-loading.md)

The Go package loads MaaFramework dynamic libraries at runtime. Installing it with `go get` does not install those libraries. Download the archive for your operating system and architecture from [MaaFramework releases](https://github.com/MaaXYZ/MaaFramework/releases); see the [installation table](../../../README.md#2-download-maaframework) and [version policy](../../../README.md#requirements).

## Library files

`Init` needs all four libraries, even if your program does not use Agents:

| Library | Windows | Linux / Android | macOS |
| --- | --- | --- | --- |
| MaaFramework | `MaaFramework.dll` | `libMaaFramework.so` | `libMaaFramework.dylib` |
| MaaToolkit | `MaaToolkit.dll` | `libMaaToolkit.so` | `libMaaToolkit.dylib` |
| MaaAgentServer | `MaaAgentServer.dll` | `libMaaAgentServer.so` | `libMaaAgentServer.dylib` |
| MaaAgentClient | `MaaAgentClient.dll` | `libMaaAgentClient.so` | `libMaaAgentClient.dylib` |

Use all four from one compatible MaaFramework release. Extract the full archive and keep its dependencies and directory layout; copying only these four files may leave their native dependencies unavailable. The compatibility contract is documented by [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init).

Desktop archives place runtime libraries in `bin/`. For Android library placement and builds, see the [Android build guide](android.md).

## Choose a loading method

Choose one of the following methods to suit your deployment environment.

### Specify the library directory

Pass the directory containing the library files, rather than the archive root or a single library filename:

```go
if err := maa.Init(maa.WithLibDir("path/to/MaaFramework/bin")); err != nil {
	return err
}
```

An absolute path avoids dependence on the launch directory. Relative paths are resolved against the process's working directory. On Windows, review the process-wide DLL search effects in [WithLibDir](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#WithLibDir) when embedding the binding in a larger application.

### Use the working directory

Place the runtime libraries and their dependencies in the directory from which you launch the program, then select that directory explicitly:

```go
if err := maa.Init(maa.WithLibDir(".")); err != nil {
	return err
}
```

The working directory can differ from the executable's directory, especially when launched by an IDE or a service. `Init()` without `WithLibDir` uses the platform loader's current search configuration; it does not guarantee a working-directory search on every platform.

### Configure environment variables

Set the search path before launching the program, then call `maa.Init()` without a library-directory option. Use an absolute path to the extracted library directory.

Windows (PowerShell):

```powershell
$env:PATH = "C:\MaaFramework\bin;" + $env:PATH
.\your-program.exe
```

Linux (shell):

```shell
LD_LIBRARY_PATH="/absolute/path/to/MaaFramework/bin${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}" ./your-program
```

macOS (shell):

```shell
DYLD_LIBRARY_PATH="/absolute/path/to/MaaFramework/bin${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}" ./your-program
```

These paths follow the operating system's loader rules: [Windows DLL search order](https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-search-order), [Linux dlopen](https://man7.org/linux/man-pages/man3/dlopen.3.html), and [macOS dynamic library loading](https://developer.apple.com/library/archive/documentation/DeveloperTools/Conceptual/DynamicLibraries/100-Articles/DynamicLibraryUsageGuidelines.html). Protected macOS processes may have `DYLD_*` variables removed; see Apple's [runtime protections](https://developer.apple.com/library/archive/documentation/Security/Conceptual/System_Integrity_Protection_Guide/RuntimeProtections/RuntimeProtections.html). If a launcher does not preserve the variable, specify the directory with `WithLibDir`.

### Use system library directories

For centrally managed deployments, install the libraries and their dependencies in directories searched by the platform loader, and call `maa.Init()` without a library-directory option. Configure those directories using the operating system's deployment tools and loader rules linked above. Check for older MaaFramework copies in the search path when upgrading.

## Troubleshooting

If initialization reports a missing library, missing symbol, or cleanup error, follow the [FAQ](faq.md). Initialization and release contracts are documented by [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) and [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release).
