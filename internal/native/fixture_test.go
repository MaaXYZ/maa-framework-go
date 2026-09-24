//go:build (darwin || linux) && (amd64 || arm64)

package native

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fixtureLibrary describes one mandatory library used by the fixture tests.
type fixtureLibrary struct {
	name    string
	file    string
	entries []Entry
}

func fixtureLibraries() []fixtureLibrary {
	return []fixtureLibrary{
		{name: maaFrameworkName, file: getMaaFrameworkLibrary(), entries: frameworkEntries},
		{name: maaToolkitName, file: getMaaToolkitLibrary(), entries: toolkitEntries},
		{name: maaAgentServerName, file: getMaaAgentServerLibrary(), entries: agentServerEntries},
		{name: maaAgentClientName, file: getMaaAgentClientLibrary(), entries: agentClientEntries},
	}
}

// buildFixtureLibraries compiles a tiny shared library for every mandatory
// library into dir. Each library exports an empty function for every required
// symbol; omitted maps a library name to a single symbol to leave out, which
// lets tests exercise a missing symbol deterministically. The test is skipped
// when no C compiler is available.
func buildFixtureLibraries(t *testing.T, dir string, omitted map[string]string) {
	t.Helper()

	cc, err := exec.LookPath("cc")
	if err != nil {
		cc, err = exec.LookPath("gcc")
	}
	if err != nil {
		t.Skip("skipping fixture tests: no C compiler available")
	}

	for _, lib := range fixtureLibraries() {
		var source strings.Builder
		for _, entry := range lib.entries {
			if omitted[lib.name] == entry.name {
				continue
			}
			fmt.Fprintf(&source, "void %s(void) {}\n", entry.name)
		}

		srcPath := filepath.Join(dir, lib.name+".c")
		if err := os.WriteFile(srcPath, []byte(source.String()), 0o644); err != nil {
			t.Fatalf("write fixture source for %s: %v", lib.name, err)
		}

		outPath := filepath.Join(dir, lib.file)
		args := []string{"-shared", "-fPIC", "-o", outPath, srcPath}
		if runtime.GOOS == "darwin" {
			args = []string{"-dynamiclib", "-fPIC", "-o", outPath, srcPath}
		}

		output, err := exec.Command(cc, args...).CombinedOutput()
		if err != nil {
			t.Skipf("skipping fixture tests: failed to build %s: %v\n%s", lib.file, err, output)
		}
	}
}

func TestFixtureInitializeAndShutdown(t *testing.T) {
	dir := t.TempDir()
	buildFixtureLibraries(t, dir, nil)

	t.Cleanup(func() { _ = Shutdown() })

	if err := Initialize(dir); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if MaaVersion == nil {
		t.Fatal("MaaVersion was not registered")
	}
	if maaFramework == 0 || maaToolkit == 0 || maaAgentServer == 0 || maaAgentClient == 0 {
		t.Fatal("not all library handles were recorded")
	}

	if err := Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	assertAllHandlesCleared(t)
	assertFuncVarsNil(t, frameworkEntries)
	assertFuncVarsNil(t, toolkitEntries)
	assertFuncVarsNil(t, agentServerEntries)
	assertFuncVarsNil(t, agentClientEntries)
}

func TestFixtureMissingSymbolRollsBackEarlierLibraries(t *testing.T) {
	dir := t.TempDir()
	missing := toolkitEntries[len(toolkitEntries)/2].name
	buildFixtureLibraries(t, dir, map[string]string{maaToolkitName: missing})

	t.Cleanup(func() { _ = Shutdown() })

	err := Initialize(dir)
	if err == nil {
		t.Fatal("Initialize returned nil for a fixture missing a symbol")
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
	if symErr.LibraryPath != libPath(dir, getMaaToolkitLibrary()) {
		t.Errorf("LibraryPath = %q, want %q", symErr.LibraryPath, libPath(dir, getMaaToolkitLibrary()))
	}
	if symErr.Requirement == "" {
		t.Error("Requirement is empty")
	}

	// The framework library was opened and registered successfully, then rolled back.
	assertFuncVarsNil(t, frameworkEntries)
	assertFuncVarsNil(t, toolkitEntries)
	assertAllHandlesCleared(t)
	if len(loadedLibs) != 0 {
		t.Errorf("loadedLibs = %d, want 0", len(loadedLibs))
	}
}
