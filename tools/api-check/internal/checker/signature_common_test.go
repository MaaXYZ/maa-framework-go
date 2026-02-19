package checker

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

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
