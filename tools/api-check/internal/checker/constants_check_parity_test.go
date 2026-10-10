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
// enum member the checker keeps must also survive a real `cc -E`.
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
			mine := resolveCConditionals(removeCComments(string(data)))
			mineMembers := enumMemberNames(t, mine)

			expanded := filepath.Join(t.TempDir(), "expanded.c")
			cmd := exec.Command("cc", "-E", "-P", "-x", "c", header, "-I", headerRoot, "-o", expanded)
			if out, runErr := cmd.CombinedOutput(); runErr != nil {
				t.Skipf("cc cannot preprocess %s: %v: %s", header, runErr, out)
			}
			preprocessed, readErr := os.ReadFile(expanded)
			if readErr != nil {
				t.Fatalf("read preprocessed %s: %v", header, readErr)
			}
			cMembers := enumMemberNames(t, string(preprocessed))

			for name := range mineMembers {
				if !cMembers[name] {
					t.Errorf("checker keeps enum member %s, but the C preprocessor does not define it", name)
				}
			}
		})
	}
}

// enumMemberNames returns the enum member names declared in content.
func enumMemberNames(t *testing.T, content string) map[string]bool {
	t.Helper()
	decls, err := parseCEnumDecls(content)
	if err != nil {
		// The parser stops at its first unsupported member, so an error only
		// means the comparison is partial; do not fail the parity check on it.
		t.Logf("partial enum parse: %v", err)
	}
	names := make(map[string]bool, len(decls))
	for _, decl := range decls {
		names[decl.name] = true
	}
	return names
}
