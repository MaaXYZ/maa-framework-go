package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseRegistrationSources(t *testing.T, sources ...string) []*ast.File {
	t.Helper()
	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(sources))
	for i, src := range sources {
		f, err := parser.ParseFile(fset, fmt.Sprintf("registration_fixture_%d.go", i), src, 0)
		if err != nil {
			t.Fatalf("parse fixture %d: %v", i, err)
		}
		files = append(files, f)
	}
	return files
}

func writeRegistrationNativeFile(t *testing.T, root, name, body string) string {
	t.Helper()
	path := filepath.Join(root, "internal", "native", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

const registrationLibrariesSource = `package native

var (
	maaFramework   uintptr
	maaToolkit     uintptr
	maaAgentServer uintptr
	maaAgentClient uintptr
)

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}
`

func TestRegistrationLibraryEntryTablesLiveFixture(t *testing.T) {
	t.Parallel()

	root, err := detectRepoRoot()
	if err != nil {
		t.Fatalf("detectRepoRoot() error = %v", err)
	}
	nativeFiles, err := discoverNativeFiles(root)
	if err != nil {
		t.Fatalf("discoverNativeFiles() error = %v", err)
	}

	fset := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(nativeFiles["framework"]))
	for _, path := range nativeFiles["framework"] {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		parsed = append(parsed, f)
	}

	ownership, live, err := libraryEntryTables(parsed)
	if err != nil {
		t.Fatalf("libraryEntryTables() error = %v", err)
	}
	if !live {
		t.Fatal("live libraries declaration must be detected")
	}
	if len(ownership) != len(moduleOrder) {
		t.Fatalf("ownership has %d modules, want %d: %v", len(ownership), len(moduleOrder), ownership)
	}
	want := map[string]string{
		"framework":    "frameworkEntries",
		"toolkit":      "toolkitEntries",
		"agent_server": "agentServerEntries",
		"agent_client": "agentClientEntries",
	}
	for module, table := range want {
		if len(ownership[module]) != 1 || !ownership[module][table] {
			t.Fatalf("ownership[%s] = %v, want only %s", module, ownership[module], table)
		}
	}
}

// TestRegistrationLiveNativeSignaturesResolve keeps the repository's real
// native Entry tables honest: every registered symbol must resolve to a Go
// function signature in the same module.
func TestRegistrationLiveNativeSignaturesResolve(t *testing.T) {
	t.Parallel()

	root, err := detectRepoRoot()
	if err != nil {
		t.Fatalf("detectRepoRoot() error = %v", err)
	}
	nativeFiles, err := discoverNativeFiles(root)
	if err != nil {
		t.Fatalf("discoverNativeFiles() error = %v", err)
	}

	registered, goSigs, _, _, issues, err := parseGoRegistrations(nativeFiles)
	if err != nil {
		t.Fatalf("parseGoRegistrations() error = %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("live native registrations reported issues: %+v", issues)
	}
	total := 0
	for _, module := range moduleOrder {
		for name := range registered[module] {
			total++
			if _, ok := goSigs[module][name]; !ok {
				t.Errorf("[%s] registered symbol %s has no resolved Go signature", module, name)
			}
		}
	}
	if total == 0 {
		t.Fatal("no live native registrations discovered")
	}
}

func TestRegistrationLibraryEntryTablesUnconsumed(t *testing.T) {
	t.Parallel()

	src := `package native

var (
	maaFramework   uintptr
	maaToolkit     uintptr
	maaAgentServer uintptr
	maaAgentClient uintptr
)

var (
	MaaVersion func() string
	MaaOld     func() string
)

var libraries = []Library{
	{entries: frameworkEntries, handle: &maaFramework},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}

var frameworkEntries = []Entry{{ptrToFunc: &MaaVersion, name: "MaaVersion"}}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}

var obsoleteEntries = []Entry{{ptrToFunc: &MaaOld, name: "MaaOld"}}
`
	ownership, live, err := libraryEntryTables(parseRegistrationSources(t, src))
	if err != nil {
		t.Fatalf("libraryEntryTables() error = %v", err)
	}
	if !live {
		t.Fatal("libraries declaration must be reported as live")
	}
	if len(ownership["framework"]) != 1 || !ownership["framework"]["frameworkEntries"] {
		t.Fatalf("framework ownership = %v, want only frameworkEntries", ownership["framework"])
	}
	if ownership["framework"]["obsoleteEntries"] {
		t.Fatalf("unconsumed obsoleteEntries must not be owned: %v", ownership["framework"])
	}
	for module, table := range map[string]string{
		"toolkit":      "toolkitEntries",
		"agent_server": "agentServerEntries",
		"agent_client": "agentClientEntries",
	} {
		if len(ownership[module]) != 1 || !ownership[module][table] {
			t.Fatalf("ownership[%s] = %v, want only %s", module, ownership[module], table)
		}
	}
}

func TestRegistrationLibraryEntryTablesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sources []string
		wantErr string
	}{
		{
			name: "duplicate libraries declaration",
			sources: []string{
				"package native\n\nvar libraries = []Library{{handle: &maaFramework, entries: frameworkEntries}}\n",
				"package native\n\nvar libraries = []Library{{handle: &maaFramework, entries: frameworkEntries}}\n",
			},
			wantErr: "duplicate libraries declaration",
		},
		{
			name:    "libraries initializer is not a composite literal",
			sources: []string{"package native\n\nvar libraries = buildLibraries()\n"},
			wantErr: "unsupported libraries initializer",
		},
		{
			name:    "libraries declaration has no initializer",
			sources: []string{"package native\n\nvar libraries []Library\n"},
			wantErr: "unsupported libraries declaration",
		},
		{
			name:    "Library element is not a composite literal",
			sources: []string{"package native\n\nvar libraries = []Library{1}\n"},
			wantErr: "unsupported Library entry",
		},
		{
			name:    "Library entry uses unnamed fields",
			sources: []string{"package native\n\nvar libraries = []Library{{&maaFramework, frameworkEntries}}\n"},
			wantErr: "Library entries must use named fields",
		},
		{
			name: "Library field key is not an identifier",
			sources: []string{`package native

var libraries = []Library{{"handle": &maaFramework, entries: frameworkEntries}}
`},
			wantErr: "unsupported Library field",
		},
		{
			name:    "Library.entries is not a named table",
			sources: []string{"package native\n\nvar libraries = []Library{{handle: &maaFramework, entries: []Entry{}}}\n"},
			wantErr: "Library.entries must reference a named Entry table",
		},
		{
			name:    "Library handle is unknown",
			sources: []string{"package native\n\nvar libraries = []Library{{handle: &maaUnknown, entries: frameworkEntries}}\n"},
			wantErr: "unknown module handle",
		},
		{
			name:    "Library handle is not addressed",
			sources: []string{"package native\n\nvar libraries = []Library{{handle: maaFramework, entries: frameworkEntries}}\n"},
			wantErr: "unknown module handle",
		},
		{
			name: "duplicate Library module",
			sources: []string{`package native

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaFramework, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}
`},
			wantErr: "duplicate Library module framework",
		},
		{
			name: "libraries missing a module",
			sources: []string{`package native

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
}
`},
			wantErr: "libraries is missing module agent_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := libraryEntryTables(parseRegistrationSources(t, tt.sources...))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("libraryEntryTables() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRegistrationLibraryEntryTablesAbsent(t *testing.T) {
	t.Parallel()

	ownership, live, err := libraryEntryTables(parseRegistrationSources(t, "package native\n\nvar version = 1\n"))
	if err != nil {
		t.Fatalf("libraryEntryTables() error = %v", err)
	}
	if live {
		t.Fatal("absent libraries declaration must not be reported as live")
	}
	if len(ownership) != 0 {
		t.Fatalf("ownership = %v, want empty", ownership)
	}
}

func TestRegistrationParseGoRegistrationsOwnership(t *testing.T) {
	t.Parallel()

	t.Run("unconsumed Entry table does not register symbols", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		path := writeRegistrationNativeFile(t, root, "framework.go", `package native

var (
	maaFramework   uintptr
	maaToolkit     uintptr
	maaAgentServer uintptr
	maaAgentClient uintptr
)

var (
	MaaVersion func() string
	MaaOld     func() string
)

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}

var frameworkEntries = []Entry{{&MaaVersion, "MaaVersion"}}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}

var obsoleteEntries = []Entry{{&MaaOld, "MaaOld"}}
`)
		registered, goSigs, _, _, issues, err := parseGoRegistrations(map[string][]string{"framework": {path}})
		if err != nil {
			t.Fatalf("parseGoRegistrations() error = %v", err)
		}
		if len(issues) != 0 {
			t.Fatalf("expected no issues, got: %+v", issues)
		}
		if _, ok := registered["framework"]["MaaVersion"]; !ok {
			t.Fatalf("MaaVersion from the live table must be registered: %v", registered["framework"])
		}
		if _, ok := registered["framework"]["MaaOld"]; ok {
			t.Fatalf("MaaOld from an unconsumed table must not be registered: %v", registered["framework"])
		}
		if _, ok := goSigs["framework"]["MaaOld"]; ok {
			t.Fatalf("unconsumed table must not contribute signatures")
		}
	})

	t.Run("libraries referencing a missing table is an error", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		path := writeRegistrationNativeFile(t, root, "framework.go", registrationLibrariesSource)
		_, _, _, _, _, err := parseGoRegistrations(map[string][]string{"framework": {path}})
		if err == nil || !strings.Contains(err.Error(), "missing Entry table frameworkEntries") {
			t.Fatalf("parseGoRegistrations() error = %v, want missing table error", err)
		}
	})
}

func TestRegistrationDiscoverNativeFiles(t *testing.T) {
	t.Parallel()

	t.Run("discovers sources and excludes tests", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeRegistrationNativeFile(t, root, "native.go", "package native\n")
		writeRegistrationNativeFile(t, root, "framework.go", "package native\n")
		writeRegistrationNativeFile(t, root, "framework_test.go", "package native\n")
		writeRegistrationNativeFile(t, root, "fixture_test.go", "package native\n")
		writeRegistrationNativeFile(t, root, "notes.txt", "not go\n")
		writeRegistrationNativeFile(t, root, filepath.Join("sub", "nested.go"), "package native\n")

		files, err := discoverNativeFiles(root)
		if err != nil {
			t.Fatalf("discoverNativeFiles() error = %v", err)
		}
		if len(files) != len(moduleOrder) {
			t.Fatalf("module map has %d entries, want %d", len(files), len(moduleOrder))
		}
		want := []string{
			filepath.Join(root, "internal", "native", "framework.go"),
			filepath.Join(root, "internal", "native", "native.go"),
		}
		for module, got := range files {
			if len(got) != len(want) {
				t.Fatalf("%s files = %v, want %v", module, got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("%s files = %v, want %v", module, got, want)
				}
			}
		}
	})

	t.Run("empty native directory is an error", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "internal", "native"), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := discoverNativeFiles(root); err == nil || !strings.Contains(err.Error(), "no native source files") {
			t.Fatalf("discoverNativeFiles() error = %v, want empty directory error", err)
		}
	})

	t.Run("missing native directory is an error", func(t *testing.T) {
		t.Parallel()

		if _, err := discoverNativeFiles(t.TempDir()); err == nil {
			t.Fatal("discoverNativeFiles() must fail without internal/native")
		}
	})
}

func TestRegistrationCheckNativeLibraryEntries(t *testing.T) {
	t.Parallel()

	t.Run("live libraries declaration passes", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeRegistrationNativeFile(t, root, "native.go", registrationLibrariesSource)
		if err := checkNativeLibraryEntries(root); err != nil {
			t.Fatalf("checkNativeLibraryEntries() error = %v", err)
		}
	})

	t.Run("missing libraries declaration fails", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeRegistrationNativeFile(t, root, "native.go", "package native\n\nvar version = 1\n")
		if err := checkNativeLibraryEntries(root); err == nil || !strings.Contains(err.Error(), "libraries declaration not found") {
			t.Fatalf("checkNativeLibraryEntries() error = %v, want missing declaration error", err)
		}
	})

	t.Run("unsupported libraries declaration fails", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeRegistrationNativeFile(t, root, "native.go", "package native\n\nvar libraries = buildLibraries()\n")
		err := checkNativeLibraryEntries(root)
		if err == nil || !strings.Contains(err.Error(), "unsupported libraries initializer") || !strings.Contains(err.Error(), "native.go") {
			t.Fatalf("checkNativeLibraryEntries() error = %v, want unsupported initializer with path", err)
		}
	})

	t.Run("syntax error fails", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeRegistrationNativeFile(t, root, "native.go", "package native\n\nfunc broken(\n")
		if err := checkNativeLibraryEntries(root); err == nil {
			t.Fatal("checkNativeLibraryEntries() must fail on a syntax error")
		}
	})
}
