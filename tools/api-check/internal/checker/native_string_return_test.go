package checker

import (
	"fmt"
	"strings"
	"testing"
)

func TestNativeCoverageCStringReturn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cReturn   string
		goReturn  string
		goParam   string
		wantIssue bool
	}{
		{name: "string", cReturn: "const char*", goReturn: "string"},
		{name: "unsafe pointer", cReturn: "const char*", goReturn: "unsafe.Pointer"},
		{name: "byte pointer", cReturn: "const char*", goReturn: "*byte"},
		{name: "uintptr", cReturn: "const char*", goReturn: "uintptr"},
		{name: "non-pointer return", cReturn: "const char*", goReturn: "bool", wantIssue: true},
		{name: "extra return", cReturn: "const char*", goReturn: "(unsafe.Pointer, bool)", wantIssue: true},
		{name: "void pointer is not a string", cReturn: "void*", goReturn: "string", wantIssue: true},
		{name: "pointer to pointer is not a string", cReturn: "const char**", goReturn: "string", wantIssue: true},
		{name: "parameter mismatch remains rejected", cReturn: "const char*", goReturn: "unsafe.Pointer", goParam: "string", wantIssue: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			param := tc.goParam
			if param == "" {
				param = "uintptr"
			}
			imports := ""
			if strings.Contains(param+tc.goReturn, "unsafe.") {
				imports = `import "unsafe"`
			}
			headerDir, files := nativeCoverageFixture(t,
				map[string]string{"framework.h": fmt.Sprintf("MAA_FRAMEWORK_API %s MaaFoo(void* handle);\n", tc.cReturn)},
				map[string]string{"framework.go": fmt.Sprintf(`package native
%s
type Entry struct {
	ptrToFunc any
	name string
}
var MaaFoo func(%s) %s
var frameworkEntries = []Entry{{&MaaFoo, "MaaFoo"}}
`, imports, param, tc.goReturn)}, "framework")
			issues, err := checkNativeAPICoverage(headerDir, files, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantIssue {
				assertIssueContains(t, issues, "signature mismatch for MaaFoo")
			} else if len(issues) != 0 {
				t.Fatalf("unexpected issues: %+v", issues)
			}
		})
	}
}
