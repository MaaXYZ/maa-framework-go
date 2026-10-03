package checker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCHeaderFixture writes one C header into root/relPath, creating parents.
func writeCHeaderFixture(t *testing.T, root string, relPath string, body string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestParseCFunctionDecl(t *testing.T) {
	t.Parallel()

	aliases := map[string]string{
		"MaaTaskId": "int64_t",
	}

	sig, name, ok := parseCFunctionDecl("MaaTaskId MaaFooBar(const char* name, void* handle)", aliases)
	if !ok {
		t.Fatalf("expected declaration to be parsed")
	}
	if name != "MaaFooBar" {
		t.Fatalf("unexpected function name: %s", name)
	}
	if got, want := strings.Join(sig.params, ","), "cstring,ptr"; got != want {
		t.Fatalf("unexpected params: got=%q want=%q", got, want)
	}
	if got, want := strings.Join(sig.returns, ","), "int64"; got != want {
		t.Fatalf("unexpected returns: got=%q want=%q", got, want)
	}
}

func TestParseGoVarFuncSignatures(t *testing.T) {
	t.Parallel()

	const src = `package native

var (
	FuncA func(uintptr, string) bool
	FuncB func(*byte)
	FuncC, FuncD func(uint64) int64
)
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "native_sample.go", src, 0)
	if err != nil {
		t.Fatalf("parse sample go source: %v", err)
	}

	sigs := parseGoVarFuncSignatures(file)

	assertSig := func(name string, params, returns string) {
		t.Helper()
		sig, ok := sigs[name]
		if !ok {
			t.Fatalf("missing signature for %s", name)
		}
		if got := strings.Join(sig.params, ","); got != params {
			t.Fatalf("%s params mismatch: got=%q want=%q", name, got, params)
		}
		if got := strings.Join(sig.returns, ","); got != returns {
			t.Fatalf("%s returns mismatch: got=%q want=%q", name, got, returns)
		}
	}

	assertSig("FuncA", "ptr,cstring", "bool")
	assertSig("FuncB", "ptr", "")
	assertSig("FuncC", "uint64", "int64")
	assertSig("FuncD", "uint64", "int64")
}

func TestNormalizeCTypeCanonicalAliasNonConverged(t *testing.T) {
	t.Parallel()

	aliases := map[string]string{
		"A": "B",
		"B": "A",
	}

	got := normalizeCTypeCanonical("A", aliases)
	if got != "<unsupported:c-alias:A>" {
		t.Fatalf("unexpected canonical type: %s", got)
	}
}

func TestNormalizeCTypeCanonicalInt(t *testing.T) {
	t.Parallel()

	if got := normalizeCTypeCanonical("int", nil); got != "int32" {
		t.Fatalf("C int should match Go int32, got %q", got)
	}
}

func TestNormalizeCTypeCanonicalPointerLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "char value", raw: "char", want: "uint8"},
		{name: "char pointer", raw: "char*", want: "cstring"},
		{name: "const char pointer", raw: "const char*", want: "cstring"},
		{name: "char double pointer", raw: "char**", want: "ptr"},
		{name: "char triple pointer", raw: "char***", want: "ptr"},
		{name: "const char double pointer", raw: "const char**", want: "ptr"},
		{name: "bool value", raw: "MaaBool", want: "bool"},
		{name: "bool pointer", raw: "MaaBool*", want: "ptr"},
		{name: "void", raw: "void", want: "void"},
		{name: "void pointer", raw: "void*", want: "ptr"},
		{name: "int", raw: "int", want: "int32"},
		{name: "int32", raw: "int32_t", want: "int32"},
		{name: "int64", raw: "int64_t", want: "int64"},
		{name: "uint8", raw: "uint8_t", want: "uint8"},
		{name: "uint16", raw: "uint16_t", want: "uint16"},
		{name: "uint32", raw: "uint32_t", want: "uint32"},
		{name: "uint64", raw: "uint64_t", want: "uint64"},
		{name: "size_t", raw: "size_t", want: "uint64"},
		{name: "struct pointer", raw: "const struct MaaBuffer*", want: "ptr"},
		{name: "unknown named type", raw: "MaaBuffer", want: "named:MaaBuffer"},
		{name: "callback typedef", raw: "MaaEventCallback", want: "callback:MaaEventCallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeCTypeCanonical(tt.raw, nil); got != tt.want {
				t.Fatalf("normalizeCTypeCanonical(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeCTypeCanonicalAliasChains(t *testing.T) {
	t.Parallel()

	aliases := map[string]string{
		"MaaTaskId":  "int64_t",
		"MaaSize":    "size_t",
		"MaaString":  "const char*",
		"MaaFeature": "uint64_t",
	}
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "MaaTaskId", want: "int64"},
		{raw: "MaaSize", want: "uint64"},
		{raw: "MaaString", want: "cstring"},
		{raw: "MaaString*", want: "ptr"},
		{raw: "MaaTaskId*", want: "ptr"},
		{raw: "MaaFeature", want: "uint64"},
		{raw: "MaaMissing", want: "named:MaaMissing"},
	}
	for _, tt := range tests {
		if got := normalizeCTypeCanonical(tt.raw, aliases); got != tt.want {
			t.Fatalf("normalizeCTypeCanonical(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestNormalizeCTypeCanonicalAliasCyclePointerDepth(t *testing.T) {
	t.Parallel()

	aliases := map[string]string{
		"A": "B",
		"B": "A",
	}
	if got, want := normalizeCTypeCanonical("A**", aliases), "<unsupported:c-alias:A>"; got != want {
		t.Fatalf("aliased pointer cycle should stay unsupported: got=%q want=%q", got, want)
	}
}

func TestParseCTypedefAliasesFromDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeCHeaderFixture(t, dir, "MaaFramework/MaaDef.h", `#pragma once

typedef /* boolean */ uint8_t MaaBool;
typedef uint64_t MaaTaskId;

/* Function pointer typedefs are not plain aliases. */
typedef void (*MaaNotificationCallback)(const char* message, void* trans_arg);
typedef void (*MaaWeird)(void) MaaWeirdAlias;

typedef char* MaaString;
typedef const char* MaaConstString;
`)
	writeCHeaderFixture(t, dir, "MaaFramework/Instance/MaaToolkit.h", "typedef size_t MaaToolkitSize;\n")

	aliases, err := parseCTypedefAliases(dir)
	if err != nil {
		t.Fatalf("parseCTypedefAliases() error = %v", err)
	}
	want := map[string]string{
		"MaaBool":        "uint8_t",
		"MaaTaskId":      "uint64_t",
		"MaaString":      "char*",
		"MaaConstString": "const char*",
		"MaaToolkitSize": "size_t",
	}
	for name, value := range want {
		if got := aliases[name]; got != value {
			t.Errorf("alias %s = %q, want %q", name, got, value)
		}
	}
	for _, name := range []string{"MaaNotificationCallback", "MaaWeird", "MaaWeirdAlias"} {
		if _, ok := aliases[name]; ok {
			t.Errorf("function pointer typedef %s must not become an alias", name)
		}
	}
}

func TestParseCCommentRemoval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "block comment", in: "a/*x*/b", want: "a     b"},
		{name: "line comment", in: "a//x\nb", want: "a   \nb"},
		{name: "unterminated block comment", in: "a/*x", want: "a   "},
		{name: "block comment keeps newlines", in: "a/*x\ny*/b", want: "a   \n   b"},
		{name: "comment markers in string literal", in: `const char* s = "/* not a comment */";`, want: `const char* s = "/* not a comment */";`},
		{name: "comment markers in char literal", in: `char c = '/';`, want: `char c = '/';`},
		{name: "escaped quote in string literal", in: `const char* s = "a\"/*x*/";`, want: `const char* s = "a\"/*x*/";`},
		{name: "comment keeps token separation", in: "MAA_FRAMEWORK_API/*note*/void MaaFoo(void);", want: "MAA_FRAMEWORK_API        void MaaFoo(void);"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := removeCComments(tt.in); got != tt.want {
				t.Fatalf("removeCComments(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseCFunctionDeclForms(t *testing.T) {
	t.Parallel()

	aliases := map[string]string{"MaaTaskId": "int64_t"}
	tests := []struct {
		name        string
		decl        string
		wantName    string
		wantParams  string
		wantReturns string
		wantOK      bool
	}{
		{name: "void with explicit void parameter", decl: "void MaaFoo(void)", wantName: "MaaFoo", wantOK: true},
		{name: "void with empty parameter list", decl: "void MaaFoo()", wantName: "MaaFoo", wantOK: true},
		{name: "bool return and string parameter", decl: "MaaBool MaaFoo(const char* name)", wantName: "MaaFoo", wantParams: "cstring", wantReturns: "bool", wantOK: true},
		{name: "alias return and parameter", decl: "MaaTaskId MaaFoo(MaaTaskId id)", wantName: "MaaFoo", wantParams: "int64", wantReturns: "int64", wantOK: true},
		{name: "function pointer declaration is rejected", decl: "void (*MaaFoo)(void)", wantOK: false},
		{name: "missing parameter list is rejected", decl: "MaaBool MaaFoo", wantOK: false},
		{name: "trailing tokens are rejected", decl: "MaaBool MaaFoo(void) extra", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sig, name, ok := parseCFunctionDecl(tt.decl, aliases)
			if ok != tt.wantOK {
				t.Fatalf("parseCFunctionDecl(%q) ok = %v, want %v", tt.decl, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if name != tt.wantName {
				t.Fatalf("name = %q, want %q", name, tt.wantName)
			}
			if got := strings.Join(sig.params, ","); got != tt.wantParams {
				t.Fatalf("params = %q, want %q", got, tt.wantParams)
			}
			if got := strings.Join(sig.returns, ","); got != tt.wantReturns {
				t.Fatalf("returns = %q, want %q", got, tt.wantReturns)
			}
		})
	}
}

func TestParseCHeaderFunctionSignaturesFromDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeCHeaderFixture(t, dir, "MaaFramework/MaaDef.h", `#pragma once
typedef uint8_t MaaBool;
typedef uint64_t MaaSize;
typedef int64_t MaaTaskId;
`)
	writeCHeaderFixture(t, dir, "MaaFramework/Instance/MaaFramework.h", `#pragma once

// A comment between the export macro and the return type must keep the tokens
// separated instead of fusing them into MAA_FRAMEWORK_APIMaaBool.
MAA_FRAMEWORK_API/* keep the token boundary */MaaBool MaaFrameworkInit(
    const char* user_path,
    const char* default_json /* trailing comment */
);

MAA_FRAMEWORK_API void MaaFrameworkShutdown(void);

/* deprecated entry points are skipped */
MAA_FRAMEWORK_API MAA_DEPRECATED void MaaDeprecated(void);

#define MAA_FAKE MAA_FRAMEWORK_API void MaaFromMacro(void);
#define MAA_CONTINUED(x) \
    MAA_FRAMEWORK_API void MaaFromContinuation(void);
#undef MAA_CONTINUED

MAA_OTHER_API void MaaFromOtherModule(void);
`)
	writeCHeaderFixture(t, dir, "MaaFramework/Instance/MaaToolkit.h", "MAA_TOOLKIT_API MaaSize MaaToolkitCount(MaaTaskId id);\n")
	writeCHeaderFixture(t, dir, "MaaFramework/Instance/MaaAgent.h", `MAA_AGENT_SERVER_API MaaBool MaaAgentServerStart(const char* name);
MAA_AGENT_CLIENT_API MaaBool MaaAgentClientStart(const char* name);
`)

	aliases, err := parseCTypedefAliases(dir)
	if err != nil {
		t.Fatalf("parseCTypedefAliases() error = %v", err)
	}
	sigs, err := parseHeaderFunctionSignatures(dir, aliases)
	if err != nil {
		t.Fatalf("parseHeaderFunctionSignatures() error = %v", err)
	}

	assertSignature := func(module, name, params, returns string) {
		t.Helper()
		sig, ok := sigs[module][name]
		if !ok {
			t.Fatalf("%s signature for %s not found (module has %v)", module, name, sigs[module])
		}
		if got := strings.Join(sig.params, ","); got != params {
			t.Errorf("%s params = %q, want %q", name, got, params)
		}
		if got := strings.Join(sig.returns, ","); got != returns {
			t.Errorf("%s returns = %q, want %q", name, got, returns)
		}
	}
	assertSignature("framework", "MaaFrameworkInit", "cstring,cstring", "bool")
	assertSignature("framework", "MaaFrameworkShutdown", "", "")
	assertSignature("toolkit", "MaaToolkitCount", "int64", "uint64")
	assertSignature("agent_server", "MaaAgentServerStart", "cstring", "bool")
	assertSignature("agent_client", "MaaAgentClientStart", "cstring", "bool")

	for _, name := range []string{"MaaDeprecated", "MaaFromMacro", "MaaFromContinuation", "MaaFromOtherModule"} {
		for module := range sigs {
			if _, ok := sigs[module][name]; ok {
				t.Errorf("skipped declaration %s must not be exported from %s", name, module)
			}
		}
	}
}

func TestParseCHeaderFunctionSignatureErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string]string
		wantErr string
	}{
		{
			name: "exported function pointer declaration",
			headers: map[string]string{
				"MaaDef.h": "MAA_FRAMEWORK_API void (*MaaCallback)(void* trans_arg);\n",
			},
			wantErr: "unsupported exported C declaration",
		},
		{
			name: "exported declaration without parameter list",
			headers: map[string]string{
				"MaaDef.h": "MAA_FRAMEWORK_API MaaBool MaaBroken;\n",
			},
			wantErr: "unsupported exported C declaration",
		},
		{
			name: "conflicting same-module declarations",
			headers: map[string]string{
				"MaaDef.h":      "MAA_FRAMEWORK_API MaaBool MaaFoo(void);\n",
				"MaaOtherDef.h": "MAA_FRAMEWORK_API void MaaFoo(void* handle);\n",
			},
			wantErr: "conflicting C declarations for MaaFoo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for name, body := range tt.headers {
				writeCHeaderFixture(t, dir, name, body)
			}
			if _, err := parseHeaderFunctionSignatures(dir, nil); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("parseHeaderFunctionSignatures() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseCHeaderFunctionSignatureDuplicateAgreement(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	body := "MAA_FRAMEWORK_API MaaBool MaaFoo(const char* name);\n"
	writeCHeaderFixture(t, dir, "a.h", body)
	writeCHeaderFixture(t, dir, "b.h", body)
	// The same symbol name may exist independently in another module.
	writeCHeaderFixture(t, dir, "c.h", "MAA_TOOLKIT_API void MaaFoo(void* handle);\n")

	sigs, err := parseHeaderFunctionSignatures(dir, nil)
	if err != nil {
		t.Fatalf("parseHeaderFunctionSignatures() error = %v", err)
	}
	for module, wantParams := range map[string]string{"framework": "cstring", "toolkit": "ptr"} {
		sig, ok := sigs[module]["MaaFoo"]
		if !ok {
			t.Fatalf("%s MaaFoo not found", module)
		}
		if got := strings.Join(sig.params, ","); got != wantParams {
			t.Fatalf("%s MaaFoo params = %q, want %q", module, got, wantParams)
		}
	}
}

func TestParseCHeaderFunctionSignatureAliasCycle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeCHeaderFixture(t, dir, "cycle.h", `typedef B A;
typedef A B;

MAA_FRAMEWORK_API MaaBool MaaFoo(A value);
`)

	aliases, err := parseCTypedefAliases(dir)
	if err != nil {
		t.Fatalf("parseCTypedefAliases() error = %v", err)
	}
	if aliases["A"] != "B" || aliases["B"] != "A" {
		t.Fatalf("cycle aliases not extracted: %v", aliases)
	}

	sigs, err := parseHeaderFunctionSignatures(dir, aliases)
	if err != nil {
		t.Fatalf("parseHeaderFunctionSignatures() error = %v", err)
	}
	sig, ok := sigs["framework"]["MaaFoo"]
	if !ok {
		t.Fatal("MaaFoo not found")
	}
	if got, want := strings.Join(sig.params, ","), "<unsupported:c-alias:A>"; got != want {
		t.Fatalf("aliased parameter = %q, want %q", got, want)
	}
}

func TestParseGoFuncSignaturesAcrossFiles(t *testing.T) {
	t.Parallel()

	type sourceFile struct {
		name string
		src  string
	}
	tests := []struct {
		name   string
		files  []sourceFile
		want   map[string][2]string
		absent []string
	}{
		{
			name: "declaration order, inferred literal, conversion and alias",
			files: []sourceFile{{name: "framework.go", src: `package native

var MaaFoo Handler
var MaaFromLiteral = func(handle uintptr, name string) bool { return true }
var MaaFromFunc = impl
var MaaFromAlias Alias
var NotAFunc = struct{}{}

type Handler func(uintptr, string) bool
type Alias Handler

func impl(handle uintptr) bool { return true }
`}},
			want: map[string][2]string{
				"MaaFoo":         {"ptr,cstring", "bool"},
				"MaaFromLiteral": {"ptr,cstring", "bool"},
				"MaaFromFunc":    {"ptr", "bool"},
				"MaaFromAlias":   {"ptr,cstring", "bool"},
			},
			absent: []string{"NotAFunc"},
		},
		{
			name: "cross-file type and function declarations",
			files: []sourceFile{
				{name: "types.go", src: "package native\n\ntype Handler func(uintptr) bool\n"},
				{name: "impl.go", src: "package native\n\nfunc impl(handle uintptr) bool { return true }\n"},
				{name: "vars.go", src: "package native\n\nvar FromType Handler\nvar FromFunc = impl\n"},
			},
			want: map[string][2]string{
				"FromType": {"ptr", "bool"},
				"FromFunc": {"ptr", "bool"},
			},
		},
		{
			name: "type conversion works and cycles stay unresolved",
			files: []sourceFile{{name: "cycle.go", src: `package native

type A B
type B A
type Handler func() bool

var FromConversion = Handler(nil)
var Cyclic A
`}},
			want: map[string][2]string{
				"FromConversion": {"", "bool"},
			},
			absent: []string{"Cyclic"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			parsed := make([]*ast.File, 0, len(tt.files))
			for _, file := range tt.files {
				f, err := parser.ParseFile(fset, file.name, file.src, 0)
				if err != nil {
					t.Fatalf("parse %s: %v", file.name, err)
				}
				parsed = append(parsed, f)
			}

			sigs, _ := collectGoFuncSignatures(parsed, fset, "")
			for name, want := range tt.want {
				sig, ok := sigs[name]
				if !ok {
					t.Fatalf("missing signature for %s (have %v)", name, sigs)
				}
				if got := strings.Join(sig.params, ","); got != want[0] {
					t.Errorf("%s params = %q, want %q", name, got, want[0])
				}
				if got := strings.Join(sig.returns, ","); got != want[1] {
					t.Errorf("%s returns = %q, want %q", name, got, want[1])
				}
			}
			for _, name := range tt.absent {
				if sig, ok := sigs[name]; ok {
					t.Errorf("%s must stay unresolved, got %+v", name, sig)
				}
			}
		})
	}
}
