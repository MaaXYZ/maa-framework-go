//go:build (darwin || linux || windows) && (amd64 || arm64)

package native

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var (
	errFakeOpen   = errors.New("fake open failure")
	errFakeLookup = errors.New("fake symbol lookup failure")
	errFakeClose  = errors.New("fake close failure")
)

// fakePlatform is a deterministic replacement for the platform dynamic library
// primitives. It records opens/closes and can be configured to fail opening a
// path, looking up a symbol, or closing a handle.
type fakePlatform struct {
	openErrors   map[string]error
	nullOpens    map[string]bool
	symbolErrors map[string]map[string]bool
	nullSymbols  map[string]bool
	closeErr     error

	nextHandle uintptr
	nextAddr   uintptr
	pathByHand map[uintptr]string
	opened     []string
	closed     []uintptr
}

func newFakePlatform() *fakePlatform {
	return &fakePlatform{
		openErrors:   map[string]error{},
		nullOpens:    map[string]bool{},
		symbolErrors: map[string]map[string]bool{},
		nullSymbols:  map[string]bool{},
		nextHandle:   0x10000,
		nextAddr:     0x20000000,
		pathByHand:   map[uintptr]string{},
	}
}

func (f *fakePlatform) open(path string) (uintptr, error) {
	if err := f.openErrors[path]; err != nil {
		return 0, err
	}
	if f.nullOpens[path] {
		return 0, nil
	}
	f.nextHandle += 0x1000
	handle := f.nextHandle
	f.pathByHand[handle] = path
	f.opened = append(f.opened, path)
	return handle, nil
}

func (f *fakePlatform) close(handle uintptr) error {
	f.closed = append(f.closed, handle)
	if f.closeErr != nil {
		return f.closeErr
	}
	return nil
}

func (f *fakePlatform) lookup(handle uintptr, name string) (uintptr, error) {
	if f.nullSymbols[name] {
		return 0, nil
	}
	if f.symbolErrors[f.pathByHand[handle]][name] {
		return 0, errFakeLookup
	}
	f.nextAddr += 0x10
	return f.nextAddr, nil
}

// installFake replaces the platform hooks for the duration of the test and
// restores them (and clears any loaded state) during cleanup.
func installFake(t *testing.T, f *fakePlatform) {
	t.Helper()

	oldOpen, oldClose, oldLookup, oldRegister := openLibrary, unloadLibrary, lookupSymbol, registerFunc
	t.Cleanup(func() {
		_ = Shutdown()
		openLibrary, unloadLibrary, lookupSymbol, registerFunc = oldOpen, oldClose, oldLookup, oldRegister
	})

	openLibrary = f.open
	unloadLibrary = f.close
	lookupSymbol = f.lookup
	registerFunc = oldRegister
}

func libPath(dir, fileName string) string {
	return filepath.Join(dir, fileName)
}

func frameworkPath(dir string) string {
	return libPath(dir, getMaaFrameworkLibrary())
}

func assertFuncVarsNil(t *testing.T, entries []Entry) {
	t.Helper()
	for _, entry := range entries {
		value := reflect.ValueOf(entry.ptrToFunc).Elem()
		if !value.IsNil() {
			t.Errorf("function variable for %q was not cleared", entry.name)
		}
	}
}

func assertAllHandlesCleared(t *testing.T) {
	t.Helper()
	if maaFramework != 0 {
		t.Errorf("maaFramework handle was not cleared: %#x", maaFramework)
	}
	if maaToolkit != 0 {
		t.Errorf("maaToolkit handle was not cleared: %#x", maaToolkit)
	}
	if maaAgentServer != 0 {
		t.Errorf("maaAgentServer handle was not cleared: %#x", maaAgentServer)
	}
	if maaAgentClient != 0 {
		t.Errorf("maaAgentClient handle was not cleared: %#x", maaAgentClient)
	}
}

func TestInitializeMissingLibrary(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	f.openErrors[frameworkPath(dir)] = errFakeOpen
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a missing library")
	}

	var loadErr *LibraryLoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("error %v is not a *LibraryLoadError", err)
	}
	if loadErr.LibraryName != maaFrameworkName {
		t.Errorf("LibraryName = %q, want %q", loadErr.LibraryName, maaFrameworkName)
	}
	if loadErr.LibraryPath != frameworkPath(dir) {
		t.Errorf("LibraryPath = %q, want %q", loadErr.LibraryPath, frameworkPath(dir))
	}
	if !errors.Is(err, errFakeOpen) {
		t.Errorf("error does not wrap the underlying open failure: %v", err)
	}
	if len(f.opened) != 0 {
		t.Errorf("opened libraries = %v, want none", f.opened)
	}
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}
	assertAllHandlesCleared(t)

	if err := Shutdown(); err != nil {
		t.Fatalf("Shutdown after failed Initialize: %v", err)
	}
}

func TestInitializeNullLibraryHandle(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	f.nullOpens[frameworkPath(dir)] = true
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a null library handle")
	}

	var loadErr *LibraryLoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("error %v is not a *LibraryLoadError", err)
	}
	if loadErr.Err == nil {
		t.Error("LibraryLoadError has no underlying error")
	}

	// Nothing was opened, so nothing may be closed.
	if len(f.closed) != 0 {
		t.Errorf("closed handles = %v, want none", f.closed)
	}
	assertAllHandlesCleared(t)
}

func TestInitializeMissingSymbol(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	missing := frameworkEntries[len(frameworkEntries)/2].name
	f.symbolErrors[frameworkPath(dir)] = map[string]bool{missing: true}
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a missing symbol")
	}

	var symErr *SymbolLookupError
	if !errors.As(err, &symErr) {
		t.Fatalf("error %v is not a *SymbolLookupError", err)
	}
	if symErr.LibraryName != maaFrameworkName {
		t.Errorf("LibraryName = %q, want %q", symErr.LibraryName, maaFrameworkName)
	}
	if symErr.LibraryPath != frameworkPath(dir) {
		t.Errorf("LibraryPath = %q, want %q", symErr.LibraryPath, frameworkPath(dir))
	}
	if symErr.SymbolName != missing {
		t.Errorf("SymbolName = %q, want %q", symErr.SymbolName, missing)
	}
	if symErr.Requirement == "" {
		t.Error("Requirement is empty")
	}
	if !strings.Contains(symErr.Requirement, "latest") {
		t.Errorf("Requirement %q does not mention the latest release", symErr.Requirement)
	}
	if !errors.Is(err, errFakeLookup) {
		t.Errorf("error does not wrap the underlying lookup failure: %v", err)
	}

	// No symbol may be registered when preflight fails.
	assertFuncVarsNil(t, frameworkEntries)
	assertAllHandlesCleared(t)
	if len(f.closed) != 1 {
		t.Errorf("closed handles = %v, want exactly the just-opened handle", f.closed)
	}
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}

	if err := Shutdown(); err != nil {
		t.Fatalf("Shutdown after failed Initialize: %v", err)
	}
	if len(f.closed) != 1 {
		t.Errorf("Shutdown closed handles after load already rolled them back: %v", f.closed)
	}
}

func TestInitializeMissingSymbolNullAddress(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	name := frameworkEntries[0].name
	f.nullSymbols[name] = true
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a null symbol address")
	}

	var symErr *SymbolLookupError
	if !errors.As(err, &symErr) {
		t.Fatalf("error %v is not a *SymbolLookupError", err)
	}
	if symErr.SymbolName != name {
		t.Errorf("SymbolName = %q, want %q", symErr.SymbolName, name)
	}
	if symErr.Err == nil || !strings.Contains(symErr.Err.Error(), "null") {
		t.Errorf("underlying error %v does not describe the null address", symErr.Err)
	}

	assertFuncVarsNil(t, frameworkEntries)
	assertAllHandlesCleared(t)
	if len(f.closed) != 1 {
		t.Errorf("closed handles = %v, want exactly the just-opened handle", f.closed)
	}
}

func TestInitializeMissingSymbolJoinsCloseFailure(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	missing := frameworkEntries[0].name
	f.symbolErrors[frameworkPath(dir)] = map[string]bool{missing: true}
	f.closeErr = errFakeClose
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a missing symbol")
	}

	var symErr *SymbolLookupError
	if !errors.As(err, &symErr) {
		t.Fatalf("error %v is not a *SymbolLookupError", err)
	}
	if !errors.Is(err, errFakeLookup) {
		t.Errorf("error does not wrap the lookup failure: %v", err)
	}
	if !errors.Is(err, errFakeClose) {
		t.Errorf("error does not join the close failure: %v", err)
	}
	assertFuncVarsNil(t, frameworkEntries)
	if maaFramework == 0 || len(loadedLibs) != 1 {
		t.Fatal("failed close handle was not retained for retry")
	}
	if err := Initialize(dir); err == nil {
		t.Error("Initialize succeeded while a failed close remains")
	}
	f.closeErr = nil
	if err := Shutdown(); err != nil {
		t.Fatalf("retry Shutdown: %v", err)
	}
	assertAllHandlesCleared(t)
}

func TestInitializeLaterLibraryMissingSymbolRollsBackEarlier(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	missing := toolkitEntries[len(toolkitEntries)/2].name
	f.symbolErrors[libPath(dir, getMaaToolkitLibrary())] = map[string]bool{missing: true}
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a missing toolkit symbol")
	}

	var symErr *SymbolLookupError
	if !errors.As(err, &symErr) {
		t.Fatalf("error %v is not a *SymbolLookupError", err)
	}
	if symErr.LibraryName != maaToolkitName {
		t.Errorf("LibraryName = %q, want %q", symErr.LibraryName, maaToolkitName)
	}
	if symErr.SymbolName != missing {
		t.Errorf("SymbolName = %q, want %q", symErr.SymbolName, missing)
	}

	// The framework library was opened and registered, then rolled back.
	assertFuncVarsNil(t, frameworkEntries)
	assertFuncVarsNil(t, toolkitEntries)
	assertAllHandlesCleared(t)
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}
	if len(f.closed) != 2 {
		t.Errorf("closed handles = %v, want the toolkit and framework handles", f.closed)
	}
}

func TestInitializeLaterLibraryOpenFailureRollsBackEarlier(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	f.openErrors[libPath(dir, getMaaAgentServerLibrary())] = errFakeOpen
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for an unopenable later library")
	}

	var loadErr *LibraryLoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("error %v is not a *LibraryLoadError", err)
	}
	if loadErr.LibraryName != maaAgentServerName {
		t.Errorf("LibraryName = %q, want %q", loadErr.LibraryName, maaAgentServerName)
	}

	assertFuncVarsNil(t, frameworkEntries)
	assertFuncVarsNil(t, toolkitEntries)
	assertAllHandlesCleared(t)
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}
	if len(f.closed) != 2 {
		t.Errorf("closed handles = %v, want the toolkit and framework handles", f.closed)
	}
}

func TestInitializeRegistrationPanicRollsBack(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	installFake(t, f)

	realRegister := registerFunc
	calls := 0
	registerFunc = func(ptrToFunc any, addr uintptr) {
		if calls == 2 {
			panic("registration boom")
		}
		calls++
		realRegister(ptrToFunc, addr)
	}

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil when registration panicked")
	}
	if !strings.Contains(err.Error(), "registration boom") {
		t.Errorf("error %v does not mention the registration panic", err)
	}

	var symErr *SymbolLookupError
	if errors.As(err, &symErr) {
		t.Errorf("registration failure was reported as a symbol lookup error: %v", err)
	}

	// Entries registered before the panic must be cleared on rollback.
	assertFuncVarsNil(t, frameworkEntries)
	assertAllHandlesCleared(t)
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}
	if len(f.closed) != 1 {
		t.Errorf("closed handles = %v, want the just-opened framework handle", f.closed)
	}
}

func TestInitializeRegistrationPanicJoinsCloseFailure(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	f.closeErr = errFakeClose
	installFake(t, f)

	registerFunc = func(any, uintptr) {
		panic("registration boom")
	}

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil when registration panicked")
	}
	if !errors.Is(err, errFakeClose) {
		t.Errorf("error does not join the close failure: %v", err)
	}
	assertFuncVarsNil(t, frameworkEntries)
	if maaFramework == 0 || len(loadedLibs) != 1 {
		t.Fatal("failed close handle was not retained for retry")
	}
	f.closeErr = nil
	if err := Shutdown(); err != nil {
		t.Fatalf("retry Shutdown: %v", err)
	}
	assertAllHandlesCleared(t)
}

func TestRepeatedInitializeShutdown(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	installFake(t, f)

	for round := 1; round <= 2; round++ {
		if err := Initialize(dir); err != nil {
			t.Fatalf("round %d Initialize: %v", round, err)
		}
		if MaaVersion == nil {
			t.Fatalf("round %d: MaaVersion was not registered", round)
		}
		if len(loadedLibs) != len(libraries) {
			t.Fatalf("round %d: loadedLibs = %d, want %d", round, len(loadedLibs), len(libraries))
		}

		if err := Shutdown(); err != nil {
			t.Fatalf("round %d Shutdown: %v", round, err)
		}
		assertAllHandlesCleared(t)
		assertFuncVarsNil(t, frameworkEntries)
		assertFuncVarsNil(t, toolkitEntries)
		assertFuncVarsNil(t, agentServerEntries)
		assertFuncVarsNil(t, agentClientEntries)
		if len(loadedLibs) != 0 {
			t.Fatalf("round %d: loadedLibs = %d, want 0", round, len(loadedLibs))
		}
	}

	if len(f.opened) != 2*len(libraries) {
		t.Errorf("opened libraries = %d, want %d", len(f.opened), 2*len(libraries))
	}
	if len(f.closed) != 2*len(libraries) {
		t.Errorf("closed libraries = %d, want %d", len(f.closed), 2*len(libraries))
	}

	// Shutdown with nothing loaded must be a no-op.
	if err := Shutdown(); err != nil {
		t.Fatalf("idle Shutdown: %v", err)
	}
	if len(f.closed) != 2*len(libraries) {
		t.Errorf("idle Shutdown closed handles: %d", len(f.closed))
	}
}

func TestShutdownWithoutInitialize(t *testing.T) {
	if err := Shutdown(); err != nil {
		t.Fatalf("Shutdown without Initialize: %v", err)
	}
}

func TestInitializeBlockedAfterRetainedHandlesFromFailedLoad(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	missing := agentClientEntries[len(agentClientEntries)/2].name
	f.symbolErrors[libPath(dir, getMaaAgentClientLibrary())] = map[string]bool{missing: true}
	f.closeErr = errFakeClose
	installFake(t, f)

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a missing fourth-library symbol")
	}
	if !errors.Is(err, errFakeLookup) {
		t.Errorf("error does not wrap the lookup failure: %v", err)
	}

	// The failed fourth library must not look registered, while every opened
	// handle is retained because the closes failed.
	assertFuncVarsNil(t, agentClientEntries)
	if maaFramework == 0 || maaToolkit == 0 || maaAgentServer == 0 || maaAgentClient == 0 {
		t.Error("expected four retained handles after failed closes")
	}
	if len(loadedLibs) != 4 {
		t.Fatalf("loadedLibs = %d, want 4 retained handles", len(loadedLibs))
	}

	// Retained handles must not be mistaken for completed initialization.
	if err := Initialize(dir); err == nil {
		t.Fatal("Initialize succeeded while retained handles from a failed load remain")
	}

	f.closeErr = nil
	delete(f.symbolErrors, libPath(dir, getMaaAgentClientLibrary()))
	if err := Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	assertAllHandlesCleared(t)
	assertFuncVarsNil(t, agentClientEntries)

	if err := Initialize(dir); err != nil {
		t.Fatalf("Initialize after cleanup: %v", err)
	}
	if MaaVersion == nil {
		t.Error("MaaVersion was not registered after recovery")
	}
}

func TestInitializeRepeatedAfterSuccessIsNoOp(t *testing.T) {
	dir := t.TempDir()
	f := newFakePlatform()
	installFake(t, f)

	if err := Initialize(dir); err != nil {
		t.Fatalf("first Initialize: %v", err)
	}
	opened := len(f.opened)

	if err := Initialize(dir); err != nil {
		t.Fatalf("second Initialize: %v", err)
	}
	if len(f.opened) != opened {
		t.Errorf("second Initialize opened %d libraries, want %d", len(f.opened), opened)
	}
	if MaaVersion == nil {
		t.Error("function variables were disturbed by the second Initialize")
	}
}

func TestInitializeRelativeLibDirFormsAbsoluteLibraryPaths(t *testing.T) {
	tests := []struct {
		name   string
		libDir string
	}{
		{name: "dot", libDir: "."},
		{name: "relative subdirectory", libDir: "libs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if tt.libDir != "." {
				if err := os.Mkdir(filepath.Join(root, tt.libDir), 0o755); err != nil {
					t.Fatalf("create subdirectory: %v", err)
				}
			}
			t.Chdir(root)
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Getwd: %v", err)
			}
			wantDir := cwd
			if tt.libDir != "." {
				wantDir = filepath.Join(cwd, tt.libDir)
			}

			f := newFakePlatform()
			installFake(t, f)

			if err := Initialize(tt.libDir); err != nil {
				t.Fatalf("Initialize(%q): %v", tt.libDir, err)
			}

			if len(f.opened) != len(libraries) {
				t.Fatalf("opened = %d libraries, want %d", len(f.opened), len(libraries))
			}
			for i, lib := range libraries {
				got := f.opened[i]
				want := filepath.Join(wantDir, lib.fileName())
				if got != want {
					t.Errorf("opened[%d] = %q, want %q", i, got, want)
				}
				if !filepath.IsAbs(got) {
					t.Errorf("opened[%d] = %q is not absolute", i, got)
				}
				if got == lib.fileName() {
					t.Errorf("opened[%d] collapsed to a bare loader-search name %q", i, got)
				}
			}
		})
	}
}

func TestInitializeEmptyLibDirUsesBareLoaderSearchNames(t *testing.T) {
	f := newFakePlatform()
	installFake(t, f)

	if err := Initialize(""); err != nil {
		t.Fatalf("Initialize(\"\"): %v", err)
	}

	if len(f.opened) != len(libraries) {
		t.Fatalf("opened = %d libraries, want %d", len(f.opened), len(libraries))
	}
	for i, lib := range libraries {
		if f.opened[i] != lib.fileName() {
			t.Errorf("opened[%d] = %q, want bare name %q", i, f.opened[i], lib.fileName())
		}
	}
}
