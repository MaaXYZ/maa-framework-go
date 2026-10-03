package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGoRegistrations_UsesEntryTables(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "native.go")
	src := `package native

var (
	MaaFoo func(uintptr, string) bool
	MaaBar func()
)

var entries = []Entry{
	{&MaaFoo, "MaaFoo"},
	{&MaaBar, "MaaBarAlias"},
}
`
	if err := os.WriteFile(file, []byte(src), 0o600); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	registered, goSigs, registerLocs, declLocs, issues, err := parseGoRegistrations(map[string][]string{
		"framework": {file},
	})
	if err != nil {
		t.Fatalf("parse go registrations: %v", err)
	}

	if _, ok := registered["framework"]["MaaFoo"]; !ok {
		t.Fatalf("expected MaaFoo to be registered")
	}
	if _, ok := registered["framework"]["MaaBarAlias"]; !ok {
		t.Fatalf("expected MaaBarAlias to be registered")
	}

	fooSig, ok := goSigs["framework"]["MaaFoo"]
	if !ok {
		t.Fatalf("expected MaaFoo signature to be collected")
	}
	if got, want := strings.Join(fooSig.params, ","), "ptr,cstring"; got != want {
		t.Fatalf("unexpected MaaFoo params: got=%q want=%q", got, want)
	}
	if got, want := strings.Join(fooSig.returns, ","), "bool"; got != want {
		t.Fatalf("unexpected MaaFoo returns: got=%q want=%q", got, want)
	}

	if got := declLocs["framework"]["MaaFoo"]; got.file != filepath.Clean(file) || got.line <= 0 {
		t.Fatalf("unexpected declaration location: %+v", got)
	}
	if got := registerLocs["framework"]["MaaFoo"]; got.file != filepath.Clean(file) || got.line <= 0 {
		t.Fatalf("unexpected registration location: %+v", got)
	}

	assertIssueContains(t, issues, "[framework] Entry mismatch")
	assertIssueContains(t, issues, "var=MaaBar symbol=MaaBarAlias")
}

const nativeTestAliases = "typedef uint8_t MaaBool;\n"

// nativeCoverageFixture writes C headers under deps/include and Go native
// sources under internal/native, returning the header directory and the native
// file map handed to checkNativeAPICoverage for one module.
func nativeCoverageFixture(t *testing.T, headers, sources map[string]string, module string) (string, map[string][]string) {
	t.Helper()
	if module == "" {
		module = "framework"
	}
	root := t.TempDir()
	headerDir := filepath.Join(root, "deps", "include")
	for name, body := range headers {
		writeCHeaderFixture(t, headerDir, name, body)
	}
	nativeDir := filepath.Join(root, "internal", "native")
	files := make([]string, 0, len(sources))
	for name, body := range sources {
		path := filepath.Join(nativeDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		files = append(files, path)
	}
	return headerDir, map[string][]string{module: files}
}

func TestNativeCoverageIssueTable(t *testing.T) {
	t.Parallel()

	headerFoo := "MAA_FRAMEWORK_API MaaBool MaaFoo(void* handle, const char* name);\n"
	tests := []struct {
		name        string
		headers     map[string]string
		sources     map[string]string
		blacklist   map[string]struct{}
		wantIssues  []string
		denyIssues  []string
		wantNoIssue bool
	}{
		{
			name: "consistent signature produces no issues",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(uintptr, string) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantNoIssue: true,
		},
		{
			name: "header function without Go registration",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources:    map[string]string{"framework.go": "package native\n"},
			wantIssues: []string{"[framework] header has function but Go is not registering it: MaaFoo"},
		},
		{
			name: "Go registration absent from headers",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": "MAA_FRAMEWORK_API MaaBool MaaBar(void);\n",
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] header has function but Go is not registering it: MaaBar",
				"[framework] Go registers function not found in headers: MaaFoo",
			},
		},
		{
			name: "parameter arity mismatch",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(uintptr) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go params: [ptr]",
				"c  params: [ptr cstring]",
			},
		},
		{
			name: "parameter order mismatch",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(string, uintptr) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go params: [cstring ptr]",
				"c  params: [ptr cstring]",
			},
		},
		{
			name: "return type mismatch",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(uintptr, string) int32
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go returns: [int32]",
				"c  returns: [bool]",
			},
		},
		{
			name: "named function type cannot bypass signature checking",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

type Handler func(string, uintptr) bool
var MaaFoo Handler
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go params: [cstring ptr]",
				"c  params: [ptr cstring]",
			},
		},
		{
			name: "inferred function cannot bypass signature checking",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo = func(handle uintptr) bool { return true }
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go params: [ptr]",
				"c  params: [ptr cstring]",
			},
		},
		{
			name: "char double pointer does not match Go string",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": "MAA_FRAMEWORK_API MaaBool MaaFoo(char** out);\n",
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(string) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"go params: [cstring]",
				"c  params: [ptr]",
			},
		},
		{
			name: "char triple pointer does not match Go string",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": "MAA_FRAMEWORK_API MaaBool MaaFoo(char*** out);\n",
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(string) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] signature mismatch for MaaFoo",
				"c  params: [ptr]",
			},
		},
		{
			name: "unsupported Go parameter type expression",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func([]byte) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] MaaFoo has unsupported Go param type expression",
			},
			denyIssues: []string{"signature mismatch for MaaFoo"},
		},
		{
			name: "unsupported Go return type expression",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": "MAA_FRAMEWORK_API MaaBool MaaFoo(void);\n",
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() []byte
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			wantIssues: []string{
				"[framework] MaaFoo has unsupported Go return type expression",
			},
			denyIssues: []string{"signature mismatch for MaaFoo"},
		},
		{
			name: "blacklist suppresses a symbol signature difference",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(uintptr) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			blacklist:   map[string]struct{}{"MaaFoo": {}},
			wantNoIssue: true,
		},
		{
			name: "blacklist without a current difference is stale",
			headers: map[string]string{
				"MaaDef.h":    nativeTestAliases,
				"framework.h": headerFoo,
			},
			sources: map[string]string{"framework.go": `package native

var MaaFoo func(uintptr, string) bool
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
			blacklist:  map[string]struct{}{"MaaFoo": {}},
			wantIssues: []string{"MaaFoo: stale native exclusion (no current difference)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerDir, nativeFiles := nativeCoverageFixture(t, tt.headers, tt.sources, "")
			issues, err := checkNativeAPICoverage(headerDir, nativeFiles, tt.blacklist)
			if err != nil {
				t.Fatalf("checkNativeAPICoverage() error = %v", err)
			}
			for _, want := range tt.wantIssues {
				assertIssueContains(t, issues, want)
			}
			for _, deny := range tt.denyIssues {
				assertIssueNotContains(t, issues, deny)
			}
			if tt.wantNoIssue && len(issues) != 0 {
				t.Fatalf("expected no issues, got: %+v", issues)
			}
		})
	}
}

func TestNativeCoverageRegistrationErrors(t *testing.T) {
	t.Parallel()

	allLibraries := `package native

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

	tests := []struct {
		name    string
		sources map[string]string
		wantErr string
	}{
		{
			name: "missing declaration for registered target",
			sources: map[string]string{"framework.go": `package native

var frameworkEntries = []Entry{{&MaaMissing, "MaaMissing"}}
`},
			wantErr: `cannot resolve Go function signature for registered target "MaaMissing"`,
		},
		{
			name: "duplicate registration in one table",
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() bool
var frameworkEntries = []Entry{
	{&MaaFoo, "MaaFoo"},
	{&MaaFoo, "MaaFoo"},
}
`},
			wantErr: "duplicate registration for MaaFoo",
		},
		{
			name: "duplicate registration across files",
			sources: map[string]string{
				"a.go": `package native

var MaaFoo func() bool
var aEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`,
				"b.go": `package native

var bEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`,
			},
			wantErr: "duplicate registration for MaaFoo",
		},
		{
			name: "unsupported entry target expression",
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() bool
var frameworkEntries = []Entry{{MaaFoo, "MaaFoo"}}
`},
			wantErr: "unsupported Entry",
		},
		{
			name: "unsupported keyed entry field",
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() bool
var frameworkEntries = []Entry{{ptrToFunc: &MaaFoo, symbol: "MaaFoo"}}
`},
			wantErr: "unsupported Entry",
		},
		{
			name: "keyed entry without both fields",
			sources: map[string]string{"framework.go": `package native

var MaaFoo func() bool
var frameworkEntries = []Entry{{name: "MaaFoo"}}
`},
			wantErr: "unsupported Entry",
		},
		{
			name: "entry table initializer is not an Entry literal",
			sources: map[string]string{"framework.go": allLibraries + `
var frameworkEntries = buildEntries()
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}
`},
			wantErr: "unsupported Entry table",
		},
		{
			name: "libraries missing a module",
			sources: map[string]string{"framework.go": `package native

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
}
var frameworkEntries = []Entry{}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
`},
			wantErr: "libraries is missing module agent_client",
		},
		{
			name: "libraries with unknown module handle",
			sources: map[string]string{"framework.go": `package native

var libraries = []Library{
	{handle: &maaUnknown, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}
var frameworkEntries = []Entry{}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}
`},
			wantErr: "unknown module handle",
		},
		{
			name: "libraries with duplicate module",
			sources: map[string]string{"framework.go": `package native

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaFramework, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}
var frameworkEntries = []Entry{}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}
`},
			wantErr: "duplicate Library module framework",
		},
		{
			name:    "libraries references a missing Entry table",
			sources: map[string]string{"framework.go": allLibraries},
			wantErr: "missing Entry table frameworkEntries",
		},
		{
			name: "libraries initializer is not a composite literal",
			sources: map[string]string{"framework.go": `package native

var libraries = buildLibraries()
`},
			wantErr: "unsupported libraries initializer",
		},
		{
			name: "libraries declaration without initializer",
			sources: map[string]string{"framework.go": `package native

var libraries []Library
`},
			wantErr: "unsupported libraries declaration",
		},
		{
			name: "Library entries must use named fields",
			sources: map[string]string{"framework.go": `package native

var libraries = []Library{{&maaFramework, frameworkEntries}}
`},
			wantErr: "Library entries must use named fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerDir, nativeFiles := nativeCoverageFixture(t, map[string]string{"MaaDef.h": nativeTestAliases}, tt.sources, "")
			_, err := checkNativeAPICoverage(headerDir, nativeFiles, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("checkNativeAPICoverage() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestNativeCoverageLiveLibraryOwnership(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"MaaDef.h": nativeTestAliases,
		"framework.h": `MAA_FRAMEWORK_API MaaBool MaaCurrent(void);
MAA_FRAMEWORK_API MaaBool MaaObsolete(void);
`,
	}
	sources := map[string]string{"framework.go": `package native

var (
	maaFramework   uintptr
	maaToolkit     uintptr
	maaAgentServer uintptr
	maaAgentClient uintptr
)

var (
	MaaCurrent  func() bool
	MaaObsolete func() bool
)

var libraries = []Library{
	{handle: &maaFramework, entries: frameworkEntries},
	{handle: &maaToolkit, entries: toolkitEntries},
	{handle: &maaAgentServer, entries: agentServerEntries},
	{handle: &maaAgentClient, entries: agentClientEntries},
}

var frameworkEntries = []Entry{{&MaaCurrent, "MaaCurrent"}}
var toolkitEntries = []Entry{}
var agentServerEntries = []Entry{}
var agentClientEntries = []Entry{}

// obsoleteEntries is not referenced by libraries and must not register symbols
// even though it has the Entry slice shape.
var obsoleteEntries = []Entry{{&MaaObsolete, "MaaObsoleteAlias"}}
`}

	headerDir, nativeFiles := nativeCoverageFixture(t, headers, sources, "")
	issues, err := checkNativeAPICoverage(headerDir, nativeFiles, nil)
	if err != nil {
		t.Fatalf("checkNativeAPICoverage() error = %v", err)
	}
	assertIssueContains(t, issues, "[framework] header has function but Go is not registering it: MaaObsolete")
	assertIssueNotContains(t, issues, "MaaCurrent")
	assertIssueNotContains(t, issues, "MaaObsoleteAlias")
	assertIssueNotContains(t, issues, "Entry mismatch")
}

func TestNativeCoverageNonLiteralTargets(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"MaaDef.h":    nativeTestAliases,
		"framework.h": "MAA_FRAMEWORK_API MaaBool MaaFoo(void* handle, const char* name);\n",
	}
	tests := []struct {
		name    string
		sources map[string]string
	}{
		{
			name: "named function type declared after the variable",
			sources: map[string]string{"framework.go": `package native

var MaaFoo Handler
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}

type Handler func(uintptr, string) bool
`},
		},
		{
			name: "inferred function literal",
			sources: map[string]string{"framework.go": `package native

var MaaFoo = func(handle uintptr, name string) bool { return true }
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
		},
		{
			name: "chained named function types",
			sources: map[string]string{"framework.go": `package native

type Base func(uintptr, string) bool
type Alias Base

var MaaFoo Alias
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`},
		},
		{
			name: "multiple names share one function type",
			sources: map[string]string{"framework.go": `package native

var MaaFoo, MaaBar Handler
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}

type Handler func(uintptr, string) bool
`},
		},
		{
			name: "cross-file named type",
			sources: map[string]string{
				"types.go": `package native

type Handler func(uintptr, string) bool
`,
				"framework.go": `package native

var MaaFoo Handler
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`,
			},
		},
		{
			name: "function type alias resolves before its declaration",
			sources: map[string]string{"framework.go": `package native

var MaaFoo Alias
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}

type Alias = Handler
type Handler func(uintptr, string) bool
`},
		},
		{
			name: "cross-file function reference",
			sources: map[string]string{
				"impl.go": `package native

func impl(handle uintptr, name string) bool { return true }
`,
				"framework.go": `package native

var MaaFoo = impl
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerDir, nativeFiles := nativeCoverageFixture(t, headers, tt.sources, "")
			issues, err := checkNativeAPICoverage(headerDir, nativeFiles, nil)
			if err != nil {
				t.Fatalf("checkNativeAPICoverage() error = %v", err)
			}
			if len(issues) != 0 {
				t.Fatalf("expected resolved signature with no issues, got: %+v", issues)
			}
		})
	}
}

func TestNativeCoverageKeyedEntryRegistration(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"MaaDef.h":    nativeTestAliases,
		"framework.h": "MAA_FRAMEWORK_API MaaBool MaaFoo(void* handle, const char* name);\n",
	}
	tests := []struct {
		name       string
		source     string
		wantIssues []string
	}{
		{
			name: "keyed Entry with both fields registers the target",
			source: `package native

var MaaFoo func(uintptr, string) bool
var frameworkEntries = []Entry{{name: "MaaFoo", ptrToFunc: &MaaFoo}}
`,
		},
		{
			name: "keyed Entry symbol mismatch is reported",
			source: `package native

var MaaFoo func(uintptr, string) bool
var frameworkEntries = []Entry{{name: "MaaFooAlias", ptrToFunc: &MaaFoo}}
`,
			wantIssues: []string{"[framework] Entry mismatch", "var=MaaFoo symbol=MaaFooAlias"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerDir, nativeFiles := nativeCoverageFixture(t, headers, map[string]string{"framework.go": tt.source}, "")
			issues, err := checkNativeAPICoverage(headerDir, nativeFiles, nil)
			if err != nil {
				t.Fatalf("checkNativeAPICoverage() error = %v", err)
			}
			for _, want := range tt.wantIssues {
				assertIssueContains(t, issues, want)
			}
			if len(tt.wantIssues) == 0 && len(issues) != 0 {
				t.Fatalf("expected no issues, got: %+v", issues)
			}
		})
	}
}
