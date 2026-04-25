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
