//go:build (darwin || linux || windows) && (amd64 || arm64)

// Package native binds the MaaFramework C API for Go through purego. The four
// mandatory dynamic libraries (MaaFramework, MaaToolkit, MaaAgentServer and
// MaaAgentClient) are loaded by Initialize from an explicit directory or the
// platform loader's current search configuration, and their C functions are
// resolved into the package-level function variables named after them.
// Shutdown unloads the libraries but does not restore DLL search configuration
// changed on Windows.
//
// Deprecated C APIs are intentionally not bound. The package is not safe for
// concurrent use: Initialize and Shutdown must be serialized by the caller
// (the public maa package does so), and no function of this package may be
// called after Shutdown: successfully unloaded libraries leave their variables
// nil, and function values captured before Shutdown reference native code
// whose lifetime is no longer guaranteed.
//
// Initialize supports linux, android, darwin and windows on amd64 or arm64 and
// panics for any other GOOS it compiles on.
package native

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/ebitengine/purego"
)

// Platform primitives are indirected through package variables so tests can
// exercise failure paths deterministically without loading real shared
// libraries. Production behavior is provided by the platform-specific
// implementations.
var (
	openLibrary   = platformOpenLibrary
	unloadLibrary = platformUnloadLibrary
	lookupSymbol  = platformLookupSymbol
	registerFunc  = purego.RegisterFunc
)

// Library describes one mandatory dynamic library that is loaded during
// initialization.
type Library struct {
	name     string
	fileName func() string
	handle   *uintptr
	entries  []Entry
}

// Entry binds a package-level function variable to the C symbol name that
// must be resolved from a loaded library.
type Entry struct {
	ptrToFunc any
	name      string
}

var (
	libraries = []Library{
		{name: maaFrameworkName, fileName: getMaaFrameworkLibrary, handle: &maaFramework, entries: frameworkEntries},
		{name: maaToolkitName, fileName: getMaaToolkitLibrary, handle: &maaToolkit, entries: toolkitEntries},
		{name: maaAgentServerName, fileName: getMaaAgentServerLibrary, handle: &maaAgentServer, entries: agentServerEntries},
		{name: maaAgentClientName, fileName: getMaaAgentClientLibrary, handle: &maaAgentClient, entries: agentClientEntries},
	}
	loadedLibs []loadedLibrary
	// initialized reports that every mandatory library was loaded and its
	// symbols registered completely. It is tracked separately from
	// len(loadedLibs) so retained handles from a failed load or unload cannot
	// be mistaken for a finished initialization.
	initialized bool
)

// loadedLibrary records a successfully opened library together with the handle
// that was returned by the platform loader so Shutdown can close exactly what
// was opened.
type loadedLibrary struct {
	lib    Library
	handle uintptr
}

// Initialize loads all mandatory MaaFramework dynamic libraries from libDir
// and resolves their symbols. A nonempty libDir is treated as an explicit
// directory and resolved to an absolute path before any library filename is
// formed; an empty libDir uses the platform loader's current search configuration.
// If any library fails to open, is missing a required symbol, or cannot be
// registered, every library opened during this call is rolled back and the
// returned error describes the failure. After a complete successful load,
// repeated calls are a no-op success until Shutdown runs.
//
// On Windows, applying a nonempty directory calls SetDllDirectoryW for the
// whole process. Once that call succeeds, neither Shutdown nor initialization
// failure rollback restores the previous DLL search configuration. It remains
// in effect until process exit or another DLL search configuration change.
// A later empty-directory initialization leaves it unchanged; a nonempty one
// can replace it. For unpackaged, unprotected Win32 processes, the change can
// also affect subsequently started child processes.
func Initialize(libDir string) error {
	if initialized {
		return nil
	}
	if len(loadedLibs) != 0 {
		return errors.New("cannot initialize while previously opened libraries remain from a failed load or unload; call Shutdown to clean up first")
	}

	resolvedDir, err := resolveLibraryDir(libDir)
	if err != nil {
		return err
	}
	libDir = resolvedDir

	// An empty libDir must skip handleLibDir entirely: on Windows,
	// SetDllDirectoryW(L"") would remove the current directory from the DLL
	// search order instead of leaving the current DLL search configuration
	// unchanged.
	if libDir != "" {
		if err := handleLibDir(libDir); err != nil {
			return err
		}
	}

	for _, lib := range libraries {
		handle, err := lib.load(libDir)
		if err != nil {
			releaseErr := Shutdown()

			if releaseErr != nil {
				return fmt.Errorf("%w; error while releasing already loaded libraries: %w", err, releaseErr)
			}
			return err
		}
		loadedLibs = append(loadedLibs, loadedLibrary{lib: lib, handle: handle})
	}

	initialized = true
	return nil
}

// resolveLibraryDir treats a nonempty libDir as an explicit directory and
// resolves it to an absolute path so that values such as "." cannot collapse
// into a bare loader-search filename. An empty libDir is left empty so the
// platform loader uses its current search configuration.
func resolveLibraryDir(libDir string) (string, error) {
	if libDir == "" {
		return "", nil
	}
	absDir, err := filepath.Abs(libDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve library directory %q: %w", libDir, err)
	}
	return absDir, nil
}

// Shutdown closes every successfully opened library in reverse load order and
// clears its handle and function variables. Libraries whose close fails are
// kept so a later Shutdown can retry them. Any complete-initialization state
// is invalidated up front so a partial unload cannot leave Initialize looking
// finished.
// On Windows, the process's DLL search configuration is not restored.
func Shutdown() error {
	initialized = false

	var (
		errs   []error
		failed []loadedLibrary
	)

	for i := len(loadedLibs) - 1; i >= 0; i-- {
		loaded := loadedLibs[i]
		if err := loaded.lib.unload(loaded.handle); err != nil {
			errs = append(errs, err)
			failed = append(failed, loaded)
		}
	}

	for i, j := 0, len(failed)-1; i < j; i, j = i+1, j-1 {
		failed[i], failed[j] = failed[j], failed[i]
	}

	loadedLibs = failed

	if len(errs) > 0 {
		return fmt.Errorf("failed to release libraries: %w", errors.Join(errs...))
	}

	return nil
}

// load opens libDir/libName, preflights every required symbol, and only then
// registers the resolved addresses. On any failure the opened handle is closed
// and the library's function variables are cleared so no partial state leaks;
// if that close itself fails, the handle is retained for a later Shutdown to
// retry (retainFailedClose).
func (lib Library) load(libDir string) (uintptr, error) {
	libPath := filepath.Join(libDir, lib.fileName())

	handle, err := openLibrary(libPath)
	if err != nil || handle == 0 {
		if err == nil {
			err = errors.New("library handle is null")
		}
		return 0, &LibraryLoadError{
			LibraryName: lib.name,
			LibraryPath: libPath,
			Err:         err,
		}
	}

	addrs, err := lib.resolve(libPath, handle)
	if err != nil {
		clearFuncVars(lib.entries)

		if closeErr := unloadLibrary(handle); closeErr != nil {
			lib.retainFailedClose(handle)
			return 0, fmt.Errorf("%w; additionally failed to close library %q after symbol lookup failure: %w", err, lib.name, closeErr)
		}
		return 0, err
	}

	if err := lib.register(addrs); err != nil {
		clearFuncVars(lib.entries)

		if closeErr := unloadLibrary(handle); closeErr != nil {
			lib.retainFailedClose(handle)
			return 0, fmt.Errorf("%w; additionally failed to close library %q after registration failure: %w", err, lib.name, closeErr)
		}
		return 0, err
	}

	*lib.handle = handle

	return handle, nil
}

// retainFailedClose records a library whose close failed as if it were still
// loaded: the handle is kept in loadedLibs for a later Shutdown to retry, and
// lib.handle keeps the native handle so the package state reflects the still
// resident image. Initialize refuses to run until Shutdown releases it.
func (lib Library) retainFailedClose(handle uintptr) {
	*lib.handle = handle
	loadedLibs = append(loadedLibs, loadedLibrary{lib: lib, handle: handle})
}

// resolve preflights every required symbol before any registration occurs.
func (lib Library) resolve(libPath string, handle uintptr) ([]uintptr, error) {
	addrs := make([]uintptr, len(lib.entries))

	for i, entry := range lib.entries {
		addr, err := lookupSymbol(handle, entry.name)
		if err == nil && addr == 0 {
			// A null function address is never valid; treat it as a lookup
			// failure even if the platform reported no error.
			err = fmt.Errorf("symbol %q resolved to a null address", entry.name)
		}
		if err != nil {
			return nil, &SymbolLookupError{
				LibraryName: lib.name,
				LibraryPath: libPath,
				SymbolName:  entry.name,
				Requirement: symbolVersionRequirement,
				Err:         err,
			}
		}
		addrs[i] = addr
	}

	return addrs, nil
}

// register binds the preflighted addresses to the package function variables,
// converting a registration panic into an error so the caller can roll back.
func (lib Library) register(addrs []uintptr) (err error) {
	symbol := ""

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to register symbol %q from library %q: %v", symbol, lib.name, r)
		}
	}()

	for i, entry := range lib.entries {
		symbol = entry.name
		registerFunc(entry.ptrToFunc, addrs[i])
	}

	return nil
}

// unload closes a successfully opened handle and only then clears the handle
// and function variables.
func (lib Library) unload(handle uintptr) error {
	if handle == 0 {
		return nil
	}

	if err := unloadLibrary(handle); err != nil {
		return fmt.Errorf("failed to unload library %q: %w", lib.name, err)
	}

	if *lib.handle == handle {
		*lib.handle = 0
	}
	clearFuncVars(lib.entries)

	return nil
}

func clearFuncVars(entries []Entry) {
	for _, entry := range entries {
		clearFuncVar(entry.ptrToFunc)
	}
}

func clearFuncVar(ptr any) {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Func {
		return
	}
	val.Elem().Set(reflect.Zero(val.Elem().Type()))
}
