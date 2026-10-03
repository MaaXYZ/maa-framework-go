package checker

import (
	"fmt"
	"strings"
	"testing"
)

func TestRunCallbackReturnMutations(t *testing.T) {
	for _, tt := range []struct {
		name, methodResult, cResult, goResult, body string
		wantCode                                    int
	}{
		{"MaaBool baseline", "bool", "MaaBool", "uintptr", "return 0", 0},
		{"MaaBool bool mutation", "bool", "MaaBool", "bool", "return false", 1},
		{"void baseline", "", "void", "uintptr", "return 0", 0},
		{"void no-result mutation", "", "void", "", "", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := repoFixtureFiles()[customControllerRel]
			source = strings.Replace(source, "package maa\n", "package maa\nimport \"github.com/ebitengine/purego\"\n", 1)
			source = strings.Replace(source, "type CustomController interface { Foo() }",
				fmt.Sprintf("type CustomController interface { Foo() %s }", tt.methodResult), 1)
			source = strings.Replace(source, "func _FooAgent(handle uintptr) uintptr { return 0 }",
				fmt.Sprintf("func _FooAgent(handle uintptr) %s { %s }", tt.goResult, tt.body), 1)
			header := fmt.Sprintf("typedef uint8_t MaaBool;\nstruct MaaCustomControllerCallbacks { %s (*foo)(void* trans_arg); };\n", tt.cResult)
			dir := writeRepoFixtureWith(t, map[string]string{
				customControllerRel:                   source,
				"deps/include/" + controllerHeaderRel: header,
			})
			code, stdout, stderr := runChecker(t, dir)
			if code != tt.wantCode || stderr != "" {
				t.Fatalf("exit %d, want %d; output:\n%s%s", code, tt.wantCode, stdout, stderr)
			}
			if tt.wantCode == 0 {
				if !strings.Contains(stdout, "PASS: no inconsistencies found.") {
					t.Fatalf("valid baseline did not pass:\n%s", stdout)
				}
			} else if !strings.Contains(stdout, "callback binding ABI mismatch for Foo") || strings.Contains(stdout, "PASS") {
				t.Fatalf("callback mutation was not rejected by the CLI:\n%s", stdout)
			}
		})
	}
}
