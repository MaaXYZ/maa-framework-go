# Frequently asked questions

English | [简体中文](../../zh/guides/faq.md)

This page covers library loading and cleanup problems. For library filenames and deployment methods, start with [Runtime library loading](library-loading.md). API contracts are maintained in godoc; use the linked API entries or `go doc` in your project to read the comments for your installed binding version.

## Init fails to load a library

For `LibraryLoadError`, inspect `LibraryName`, `LibraryPath`, and the underlying `Err`:

1. Check that the attempted directory contains [all four library files](library-loading.md#library-files). With an explicit directory, pass the directory containing the libraries rather than its parent.
2. Match the release archive to the program's operating system and architecture.
3. Extract the full release archive again if the named file exists but loading still fails; a dependency may be missing. Read the underlying system error for the dependency name or other cause.
4. If using environment variables, set them before starting the process. For an IDE or service, check its working directory and inherited environment, or use an absolute `WithLibDir` path.

See [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) and [WithLibDir](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#WithLibDir) for the loading contract.

## Init cannot find a required symbol

For `SymbolLookupError`, record `LibraryPath`, `SymbolName`, `Requirement`, and the underlying `Err`. Replace all four libraries with files from one compatible release; replacing only the library named in the error can leave mixed versions. Check for older copies in environment or system search paths. Use an explicit library directory to narrow down which files are being loaded.

The header badge describes a tested baseline; choose a release using the [version policy](../../../README.md#requirements) and the compatibility requirements in [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init).

Use `errors.As` to inspect either error, including when it is wrapped. This helper logs diagnostic fields and returns the initialization error together with any error from a cleanup retry:

```go
func initialize(libDir string) error {
	if err := maa.Init(maa.WithLibDir(libDir)); err != nil {
		var loadErr *maa.LibraryLoadError
		var symbolErr *maa.SymbolLookupError
		switch {
		case errors.As(err, &loadErr):
			fmt.Fprintf(os.Stderr, "library=%s path=%s: %v\n",
				loadErr.LibraryName, loadErr.LibraryPath, loadErr.Err)
		case errors.As(err, &symbolErr):
			fmt.Fprintf(os.Stderr, "library=%s path=%s symbol=%s: %v; %s\n",
				symbolErr.LibraryName, symbolErr.LibraryPath,
				symbolErr.SymbolName, symbolErr.Err, symbolErr.Requirement)
		}
		return errors.Join(err, maa.Release())
	}
	return nil
}
```

Import `errors`, `fmt`, `os`, and `github.com/MaaXYZ/maa-framework-go/v4`. After successful initialization, the caller manages object cleanup and the final `Release`.

## Initialization or release reports an unload error

Keep the complete error, including any cleanup error. Call `Release` again and check its result before retrying `Init`; do not continue making native calls after a partial unload. If cleanup continues to fail, retain the error for a bug report and restart the process before retrying. Failure cleanup and retained handles are documented by [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) and [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release).

## Release returns ErrLibraryInUse

Recognize this error with `errors.Is(err, maa.ErrLibraryInUse)`, then check the objects and Agent Server used by your program:

1. Stop new submissions and wait for active calls, jobs, and callbacks to finish.
2. Destroy owning `AgentClient` and `Tasker` objects before destroying resources or controllers they retain. Also destroy any remaining owning `Resource`, `Controller`, or `PortalHelper` objects. Check every `Destroy` result; fix the reported error before retrying `Release`.
3. If you started an Agent Server without detaching it, follow the shutdown sequence in [AgentServerJoin](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerJoin), then retry `Release`. A return from `Join` alone does not complete server cleanup.
4. If you called `AgentServerDetach`, follow the next section.

Ownership and destruction rules are documented in the [package overview](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#pkg-overview), [Tasker.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Tasker.Destroy), [Resource.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Resource.Destroy), [Controller.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Controller.Destroy), [AgentClient.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentClient.Destroy), and [PortalHelper.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#PortalHelper.Destroy). The release guard is documented by [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release).

## Release is blocked after AgentServerDetach

Keep the libraries loaded until process exit. Calling `Join` or `ShutDown` afterward cannot make library unloading safe. For a program that needs to release libraries, remove `Detach` and restart the process; select a shutdown sequence from [AgentServerJoin](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerJoin).

You do not need `Detach` merely to run the server in the background. Review [AgentServerStartUp](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerStartUp) and [AgentServerDetach](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerDetach) when choosing the server lifetime.

## Reporting an unresolved problem

Include the binding version, MaaFramework release, operating system and architecture, loading method, complete error, and a minimal reproduction. Remove credentials or other private data from logs. Contribution instructions are in [CONTRIBUTING.md](../../../CONTRIBUTING.md).
