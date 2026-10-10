package checker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveCConditionals_MatchesCHeaders cross-checks the checker's own
// conditional handling against the system C preprocessor on the headers that
// ship with the local MaaFramework release.
//
// resolveCConditionals exists because the release headers guard an enum member
// with `#if defined(__cplusplus)` to attach a C++-only attribute. Reducing that
// guard by hand is easy to get wrong, so the parity check is the contract: every
// enum member the checker keeps must have the same value after a real `cc -E`.
//
// Only enum members are compared. Object-like macros are expanded away by the
// preprocessor, so their identifiers are not a meaningful oracle here.
func TestResolveCConditionals_MatchesCHeaders(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("cc is required to compare against the C preprocessor")
	}
	headerRoot := filepath.Join("..", "..", "..", "..", "deps", "include")
	var headers []string
	walkErr := filepath.Walk(headerRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".h") {
			headers = append(headers, path)
		}
		return nil
	})
	if walkErr != nil || len(headers) == 0 {
		t.Skipf("no headers under %s: %v", headerRoot, walkErr)
	}

	for _, header := range headers {
		t.Run(filepath.Base(header), func(t *testing.T) {
			t.Parallel()
			data, readErr := os.ReadFile(header)
			if readErr != nil {
				t.Fatalf("read %s: %v", header, readErr)
			}
			mine := resolveCContentForParity(t, string(data))
			mineMembers := enumMemberValues(t, mine)

			expanded := filepath.Join(t.TempDir(), "expanded.c")
			cmd := exec.Command("cc", "-E", "-P", "-x", "c", header, "-I", headerRoot, "-o", expanded)
			if out, runErr := cmd.CombinedOutput(); runErr != nil {
				t.Skipf("cc cannot preprocess %s: %v: %s", header, runErr, out)
			}
			preprocessed, readErr := os.ReadFile(expanded)
			if readErr != nil {
				t.Fatalf("read preprocessed %s: %v", header, readErr)
			}
			cMembers := enumMemberValues(t, string(preprocessed))

			for name, value := range mineMembers {
				if cValue, ok := cMembers[name]; !ok {
					t.Errorf("checker keeps enum member %s, but the C preprocessor does not define it", name)
				} else if value != cValue {
					t.Errorf("enum member %s: checker value %s, C preprocessor value %s", name, value, cValue)
				}
			}
		})
	}
}

// TestResolveCConditionals_MatchesCPreprocessor does not depend on release
// headers. Its C and C++ branches deliberately differ in both names and values,
// so selecting the wrong branch cannot pass a name-only comparison.
func TestResolveCConditionals_MatchesCPreprocessor(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("cc is required to compare against the C preprocessor")
	}
	content := `enum Fixture {
    MaaFixture_Start = 1,
#if defined(__cplusplus)
    MaaFixture_Provider = 99,
    MaaFixture_CppOnly = 100,
#else
    MaaFixture_Provider = 2,
    MaaFixture_COnly,
#endif
#if !defined __cplusplus
    MaaFixture_Negated,
#elif !defined(__cplusplus)
    MaaFixture_Unreachable = 99,
#else
    MaaFixture_Unreachable = 100,
#endif
#ifndef __cplusplus
#if !defined(__cplusplus)
    MaaFixture_Nested,
#endif
#else
    MaaFixture_Nested = 99,
#endif
#ifdef __cplusplus
    MaaFixture_Last = 99,
#else
    MaaFixture_Last,
#endif
};
`
	header := filepath.Join(t.TempDir(), "fixture.h")
	if err := os.WriteFile(header, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cc", "-E", "-P", "-x", "c", header)
	expanded, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cc cannot preprocess fixture: %v: %s", err, expanded)
	}
	mine := enumMemberValues(t, resolveCContentForParity(t, content))
	want := enumMemberValues(t, string(expanded))
	if len(mine) != len(want) {
		t.Errorf("checker enum members = %v, C preprocessor enum members = %v", mine, want)
	}
	for name, value := range want {
		if got := mine[name]; got != value {
			t.Errorf("enum member %s: checker value %q, C preprocessor value %s", name, got, value)
		}
	}
}

func resolveCContentForParity(t *testing.T, content string) string {
	t.Helper()
	resolved, err := resolveCConditionals(removeCComments(content))
	if err != nil {
		t.Fatalf("resolve C conditionals: %v", err)
	}
	return resolved
}

// enumMemberValues returns the exact integer values of enum members in content.
// A parsing or evaluation failure must fail the oracle instead of comparing a
// partial inventory.
func enumMemberValues(t *testing.T, content string) map[string]string {
	t.Helper()
	decls, err := parseCEnumDecls(content)
	if err != nil {
		t.Fatalf("parse enum members: %v", err)
	}
	if len(decls) == 0 {
		// Platform-specific export macros in headers without enums are outside
		// this oracle's scope and need no numeric evaluation.
		return map[string]string{}
	}
	env, err := evaluateCConstSources([]cConstSource{{path: "parity.h", content: content}})
	if err != nil {
		t.Fatalf("evaluate enum members: %v", err)
	}
	values := make(map[string]string, len(decls))
	for _, decl := range decls {
		value, ok := env.values[decl.name]
		if !ok {
			t.Fatalf("enum member %s has no value: %s", decl.name, env.failures[decl.name])
		}
		values[decl.name] = value.ExactString()
	}
	return values
}
