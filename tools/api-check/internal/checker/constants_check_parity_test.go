package checker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestResolveCConditionals_MatchesCHeaders cross-checks the checker's own
// conditional handling against the system C preprocessor on the headers that
// ship with the local MaaFramework release.
//
// resolveCConditionals exists because the release headers guard an enum member
// with `#if defined(__cplusplus)` to attach a C++-only attribute. Reducing that
// guard by hand is easy to get wrong, so the parity check is the contract: the
// checker and a real `cc -E` must keep the same enum names and values directly
// declared in each header. Enum members from included headers are not compared
// against a parser that only reads that header's source.
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
			cmd := exec.Command("cc", "-E", "-x", "c", header, "-I", headerRoot, "-o", expanded)
			if out, runErr := cmd.CombinedOutput(); runErr != nil {
				t.Skipf("cc cannot preprocess %s: %v: %s", header, runErr, out)
			}
			preprocessed, readErr := os.ReadFile(expanded)
			if readErr != nil {
				t.Fatalf("read preprocessed %s: %v", header, readErr)
			}
			cMembers := preprocessedHeaderEnumValues(t, string(preprocessed), header)

			for _, difference := range enumMemberDifferences(mineMembers, cMembers) {
				t.Error(difference)
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
	for _, difference := range enumMemberDifferences(mine, want) {
		t.Error(difference)
	}
}

func TestResolveCConditionals_ParityWithIncludedHeader(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("cc is required to compare against the C preprocessor")
	}
	dir := t.TempDir()
	dependency := filepath.Join(dir, "dependency.h")
	if err := os.WriteFile(dependency, []byte("enum Dependency { MaaDependency_Value = 7 };\n"), 0600); err != nil {
		t.Fatal(err)
	}
	content := `#include "dependency.h"
enum Fixture {
    MaaFixture_First = 1,
    MaaFixture_Second = 2,
};
`
	header := filepath.Join(dir, "fixture with spaces.h")
	if err := os.WriteFile(header, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	expanded := filepath.Join(dir, "expanded.c")
	cmd := exec.Command("cc", "-E", "-x", "c", header, "-I", dir, "-o", expanded)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cc cannot preprocess fixture: %v: %s", err, out)
	}
	preprocessed, err := os.ReadFile(expanded)
	if err != nil {
		t.Fatal(err)
	}
	want := preprocessedHeaderEnumValues(t, string(preprocessed), header)
	if _, included := want["MaaDependency_Value"]; included || len(want) != 2 {
		t.Fatalf("direct-header enum members = %v, want only the two fixture members", want)
	}
	mine := enumMemberValues(t, resolveCContentForParity(t, content))
	if differences := enumMemberDifferences(mine, want); len(differences) != 0 {
		t.Fatalf("matching direct-header inventories differ: %v", differences)
	}

	// The release-header oracle must detect a dropped member, even when every
	// member that remains in the checker still has the correct value.
	delete(mine, "MaaFixture_Second")
	differences := enumMemberDifferences(mine, want)
	if len(differences) != 1 || differences[0] != "C preprocessor defines enum member MaaFixture_Second (value 2), but the checker does not keep it" {
		t.Fatalf("missing-member differences = %v, want the dropped fixture member", differences)
	}
}

func TestPreprocessedHeaderContents_LineMarkers(t *testing.T) {
	t.Parallel()
	header := filepath.Join(t.TempDir(), `quoted" header.h`)
	dependency := filepath.Join(filepath.Dir(header), "dependency.h")
	uncleanHeader := filepath.Dir(header) + string(filepath.Separator) + "nested" + string(filepath.Separator) + ".." + string(filepath.Separator) + filepath.Base(header)
	content := fmt.Sprintf(`# 0 %s
enum First { MaaFixture_First = 1 };
# 1 %s 1
enum Dependency { MaaDependency_Value = 7 };
#line 8 %s
enum Second { MaaFixture_Second = MaaDependency_Value };
#pragma once
`, strconv.Quote(header), strconv.Quote(dependency), strconv.Quote(uncleanHeader))
	full, direct, err := preprocessedHeaderContents(content, header)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(full, "#") || !strings.Contains(full, "MaaDependency_Value = 7") {
		t.Fatalf("full content does not retain dependency declarations without directives:\n%s", full)
	}
	if strings.Contains(direct, "MaaDependency_Value = 7") || !strings.Contains(direct, "MaaFixture_First = 1") || !strings.Contains(direct, "MaaFixture_Second = MaaDependency_Value") {
		t.Fatalf("direct content does not follow decoded, normalized line-marker paths:\n%s", direct)
	}
	values := preprocessedHeaderEnumValues(t, content, header)
	if len(values) != 2 || values["MaaFixture_First"] != "1" || values["MaaFixture_Second"] != "7" {
		t.Fatalf("direct-header enum values = %v, want First=1 and Second=7", values)
	}
}

func resolveCContentForParity(t *testing.T, content string) string {
	t.Helper()
	resolved, err := resolveCConditionals(removeCComments(spliceCLineContinuations(content)))
	if err != nil {
		t.Fatalf("resolve C conditionals: %v", err)
	}
	return resolved
}

// enumMemberDifferences compares names in both directions as well as exact
// values, so a missing checker member cannot pass a partial inventory check.
func enumMemberDifferences(mine, want map[string]string) []string {
	var differences []string
	for name, value := range mine {
		if cValue, ok := want[name]; !ok {
			differences = append(differences, fmt.Sprintf("checker keeps enum member %s, but the C preprocessor does not define it", name))
		} else if value != cValue {
			differences = append(differences, fmt.Sprintf("enum member %s: checker value %s, C preprocessor value %s", name, value, cValue))
		}
	}
	for name, value := range want {
		if _, ok := mine[name]; !ok {
			differences = append(differences, fmt.Sprintf("C preprocessor defines enum member %s (value %s), but the checker does not keep it", name, value))
		}
	}
	sort.Strings(differences)
	return differences
}

// preprocessedHeaderEnumValues evaluates the full expansion so included enum
// dependencies remain available, then keeps only names declared by header.
func preprocessedHeaderEnumValues(t *testing.T, content, header string) map[string]string {
	t.Helper()
	full, direct, err := preprocessedHeaderContents(content, header)
	if err != nil {
		t.Fatal(err)
	}
	decls, err := parseCEnumDecls(direct)
	if err != nil {
		t.Fatalf("parse direct-header enum members: %v", err)
	}
	allValues := enumMemberValues(t, full)
	values := make(map[string]string, len(decls))
	for _, decl := range decls {
		value, ok := allValues[decl.name]
		if !ok {
			t.Fatalf("direct-header enum member %s is missing from the full expansion", decl.name)
		}
		values[decl.name] = value
	}
	return values
}

var cPreprocessorLineMarkerRe = regexp.MustCompile(`^#\s*(?:line\s+)?[0-9]+\s+("(?:\\.|[^"\\])*")`)

// preprocessedHeaderContents removes all directives while retaining the source
// attribution from line markers. Quoted filenames follow C/Go string escaping.
func preprocessedHeaderContents(content, header string) (full, direct string, err error) {
	header, err = filepath.Abs(header)
	if err != nil {
		return "", "", fmt.Errorf("normalize header path: %w", err)
	}
	var fullContent, directContent strings.Builder
	currentHeader := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if match := cPreprocessorLineMarkerRe.FindStringSubmatch(trimmed); match != nil {
				filename, unquoteErr := strconv.Unquote(match[1])
				if unquoteErr != nil {
					return "", "", fmt.Errorf("decode preprocessor filename: %w", unquoteErr)
				}
				currentHeader, err = filepath.Abs(filename)
				if err != nil {
					return "", "", fmt.Errorf("normalize preprocessor filename: %w", err)
				}
			}
			continue
		}
		fullContent.WriteString(line)
		fullContent.WriteByte('\n')
		if currentHeader == header {
			directContent.WriteString(line)
			directContent.WriteByte('\n')
		}
	}
	return fullContent.String(), directContent.String(), nil
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
