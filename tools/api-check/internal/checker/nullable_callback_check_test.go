package checker

import (
	"fmt"
	"strings"
	"testing"
)

const nullableIOSCallbackHeader = `typedef uint8_t MaaBool;
typedef int32_t MaaIOSConnectionMethod;
typedef MaaBool(MAA_CALL* MaaToolkitIOSDeviceCancelCallback)(void* userdata);
MAA_TOOLKIT_API MaaBool MaaToolkitIOSDeviceFindWithOptions(
	MaaIOSConnectionMethod connection_method,
	const char* options_json,
	MaaToolkitIOSDeviceCancelCallback cancel_callback,
	void* userdata,
	void* list);
`

const nullableIOSCallbackNative = `
type MaaToolkitIOSDeviceCancelCallback func(userdata uintptr) uintptr
type MaaToolkitIOSDeviceCancelCallbackPtr uintptr

var MaaToolkitIOSDeviceFindWithOptions func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr, uintptr, uintptr) bool
`

const nullableIOSCallbackRoot = `package maa

import "github.com/ebitengine/purego"

var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)

func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uintptr) uintptr { return 0 }
`

func TestRunNullableIOSCallbackMutations(t *testing.T) {
	tests := []struct {
		name                 string
		nativeOld, nativeNew string
		headerOld, headerNew string
		rootOld, rootNew     string
		wantIssues           []string
	}{
		{name: "shared nullable callback pointer preserves ABI coverage"},
		{
			name:       "bare uintptr cannot bypass a callback parameter",
			nativeOld:  "func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr,",
			nativeNew:  "func(int32, string, uintptr,",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "another named pointer cannot bypass a callback parameter",
			nativeOld:  "MaaToolkitIOSDeviceCancelCallbackPtr",
			nativeNew:  "OtherCallbackPtr",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "a type based on the approved pointer cannot bypass a callback parameter",
			nativeOld:  "var MaaToolkitIOSDeviceFindWithOptions func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr,",
			nativeNew:  "type OtherCallbackPtr MaaToolkitIOSDeviceCancelCallbackPtr\nvar MaaToolkitIOSDeviceFindWithOptions func(int32, string, OtherCallbackPtr,",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "a type alias of the approved pointer cannot bypass a callback parameter",
			nativeOld:  "var MaaToolkitIOSDeviceFindWithOptions func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr,",
			nativeNew:  "type OtherCallbackPtr = MaaToolkitIOSDeviceCancelCallbackPtr\nvar MaaToolkitIOSDeviceFindWithOptions func(int32, string, OtherCallbackPtr,",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "direct Go callbacks cannot replace the nullable pointer",
			nativeOld:  "func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr,",
			nativeNew:  "func(int32, string, MaaToolkitIOSDeviceCancelCallback,",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "another named callback cannot replace the nullable pointer",
			nativeOld:  "func(int32, string, MaaToolkitIOSDeviceCancelCallbackPtr,",
			nativeNew:  "func(int32, string, OtherCallback,",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "32-bit pointer representation is rejected",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallbackPtr uintptr",
			nativeNew:  "type MaaToolkitIOSDeviceCancelCallbackPtr uint32",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "fixed-width integers do not establish a raw callback pointer",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallbackPtr uintptr",
			nativeNew:  "type MaaToolkitIOSDeviceCancelCallbackPtr uint64",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "data pointers do not establish a raw callback pointer",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallbackPtr uintptr",
			nativeNew:  "type MaaToolkitIOSDeviceCancelCallbackPtr *byte",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "the pointer must remain a defined type",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallbackPtr uintptr",
			nativeNew:  "type MaaToolkitIOSDeviceCancelCallbackPtr = uintptr",
			wantIssues: []string{"nullable callback pointer must be a defined uintptr type"},
		},
		{
			name:       "the pointer declaration is required",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallbackPtr uintptr",
			nativeNew:  "",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "the named Go callback declaration is required",
			nativeOld:  "type MaaToolkitIOSDeviceCancelCallback func(userdata uintptr) uintptr",
			nativeNew:  "",
			wantIssues: []string{"C callback typedef missing in Go: MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "the named Go callback parameter width is checked",
			nativeOld:  "func(userdata uintptr) uintptr",
			nativeNew:  "func(userdata uint32) uintptr",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "the named Go callback arity is checked",
			nativeOld:  "func(userdata uintptr) uintptr",
			nativeNew:  "func(userdata, extra uintptr) uintptr",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "the named Go callback return width is checked",
			nativeOld:  "func(userdata uintptr) uintptr",
			nativeNew:  "func(userdata uintptr) bool",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "the C callback typedef is required",
			headerOld:  "typedef MaaBool(MAA_CALL* MaaToolkitIOSDeviceCancelCallback)(void* userdata);",
			headerNew:  "",
			wantIssues: []string{"Go callback typedef absent from C: MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "C callback parameter width changes remain visible",
			headerOld:  "(void* userdata);",
			headerNew:  "(uint32_t userdata);",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback", "callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "C callback arity changes remain visible",
			headerOld:  "(void* userdata);",
			headerNew:  "(void* userdata, void* extra);",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback", "callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "C callback return changes remain visible",
			headerOld:  "typedef MaaBool(MAA_CALL*",
			headerNew:  "typedef int32_t(MAA_CALL*",
			wantIssues: []string{"callback ABI mismatch for MaaToolkitIOSDeviceCancelCallback", "callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "a different C callback does not match the raw pointer",
			headerOld:  "MaaToolkitIOSDeviceCancelCallback cancel_callback",
			headerNew:  "MaaOtherCallback cancel_callback",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "the raw callback pointer cannot become an untyped data pointer",
			headerOld:  "MaaToolkitIOSDeviceCancelCallback cancel_callback",
			headerNew:  "void* cancel_callback",
			wantIssues: []string{"signature mismatch for MaaToolkitIOSDeviceFindWithOptions"},
		},
		{
			name:       "the raw pointer cannot be used by another API",
			nativeOld:  "MaaToolkitIOSDeviceFindWithOptions",
			nativeNew:  "MaaToolkitOtherFindWithOptions",
			headerOld:  "MaaToolkitIOSDeviceFindWithOptions",
			headerNew:  "MaaToolkitOtherFindWithOptions",
			wantIssues: []string{"signature mismatch for MaaToolkitOtherFindWithOptions"},
		},
		{
			name:       "missing root trampoline is rejected",
			rootOld:    "func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uintptr) uintptr { return 0 }",
			rootNew:    "",
			wantIssues: []string{"missing callback trampoline for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "actual trampoline parameter width is checked",
			rootOld:    "func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uintptr)",
			rootNew:    "func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uint32)",
			wantIssues: []string{"callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "actual trampoline arity is checked",
			rootOld:    "func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uintptr)",
			rootNew:    "func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata, extra uintptr)",
			wantIssues: []string{"callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "actual trampoline return width is checked",
			rootOld:    "uintptr { return 0 }",
			rootNew:    "bool { return false }",
			wantIssues: []string{"callback trampoline ABI mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "missing NewCallback registration is rejected",
			rootOld:    "var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)",
			rootNew:    "",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "an unrelated NewCallback target with the same ABI is rejected",
			rootOld:    "purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)",
			rootNew:    "purego.NewCallback(otherAgent)\nfunc otherAgent(userdata uintptr) uintptr { return 0 }",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "a callback closure cannot replace the shared trampoline",
			rootOld:    "purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)",
			rootNew:    "purego.NewCallback(func(userdata uintptr) uintptr { return 0 })",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "NewCallback must come from purego",
			rootOld:    "github.com/ebitengine/purego",
			rootNew:    "example.com/purego",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "a second NewCallback allocation is rejected",
			rootOld:    "var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)",
			rootNew:    "var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)\nfunc registerAgain() { purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent) }",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
		{
			name:       "a function-local NewCallback is not a shared registration",
			rootOld:    "var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)",
			rootNew:    "var iosDeviceCancelCallback uintptr\nfunc register() { iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent) }",
			wantIssues: []string{"nullable callback binding mismatch for MaaToolkitIOSDeviceCancelCallback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			native := nullableIOSCallbackNative
			header := nullableIOSCallbackHeader
			root := nullableIOSCallbackRoot
			if tt.nativeOld != "" {
				native = strings.ReplaceAll(native, tt.nativeOld, tt.nativeNew)
			}
			if tt.headerOld != "" {
				header = strings.ReplaceAll(header, tt.headerOld, tt.headerNew)
			}
			if tt.rootOld != "" {
				root = strings.ReplaceAll(root, tt.rootOld, tt.rootNew)
			}
			files := repoFixtureFiles()
			toolkit := files["internal/native/toolkit.go"] + native
			symbol := "MaaToolkitIOSDeviceFindWithOptions"
			if tt.nativeOld == symbol {
				symbol = tt.nativeNew
			}
			toolkit = strings.Replace(toolkit, "var toolkitEntries = []Entry{", fmt.Sprintf("var toolkitEntries = []Entry{{&%s, %q},", symbol, symbol), 1)
			dir := writeRepoFixtureWith(t, map[string]string{
				"internal/native/toolkit.go": toolkit,
				"deps/include/ios.h":         header,
				"toolkit_ios_device.go":      root,
			})
			code, stdout, stderr := runChecker(t, dir)
			wantCode := 0
			if len(tt.wantIssues) != 0 {
				wantCode = 1
			}
			if code != wantCode || stderr != "" {
				t.Fatalf("exit %d, want %d; output:\n%s%s", code, wantCode, stdout, stderr)
			}
			for _, want := range tt.wantIssues {
				if !strings.Contains(stdout, want) {
					t.Errorf("missing issue %q in output:\n%s", want, stdout)
				}
			}
			if wantCode == 0 && !strings.Contains(stdout, "PASS: no inconsistencies found.") {
				t.Fatalf("valid nullable callback did not pass:\n%s", stdout)
			}
		})
	}
}

func TestNativeParamTypesMatchNullableCallbackScope(t *testing.T) {
	t.Parallel()
	callback := "callback:" + iosDeviceCancelCallbackName
	goTypes := []string{"int32", "cstring", iosDeviceCancelCallbackMarker, "ptr", "ptr"}
	cTypes := []string{"int32", "cstring", callback, "ptr", "ptr"}
	tests := []struct {
		name, module, function string
		goTypes, cTypes        []string
		want                   bool
	}{
		{"approved slot", "toolkit", "MaaToolkitIOSDeviceFindWithOptions", goTypes, cTypes, true},
		{"wrong library", "framework", "MaaToolkitIOSDeviceFindWithOptions", goTypes, cTypes, false},
		{
			name:     "wrong parameter position",
			module:   "toolkit",
			function: "MaaToolkitIOSDeviceFindWithOptions",
			goTypes:  []string{iosDeviceCancelCallbackMarker, "cstring", "int32", "ptr", "ptr"},
			cTypes:   []string{callback, "cstring", "int32", "ptr", "ptr"},
		},
		{"wrong API", "toolkit", "MaaToolkitOtherFind", []string{iosDeviceCancelCallbackMarker}, []string{callback}, false},
		{"unrelated callbacks remain exact", "framework", "MaaRegister", []string{"callback:MaaEventCallback"}, []string{"callback:MaaEventCallback"}, true},
		{"unrelated callbacks still reject pointers", "framework", "MaaRegister", []string{"ptr"}, []string{"callback:MaaEventCallback"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nativeParamTypesMatch(tt.module, tt.function, tt.goTypes, tt.cTypes); got != tt.want {
				t.Fatalf("nativeParamTypesMatch(%q, %q, %v, %v) = %v, want %v", tt.module, tt.function, tt.goTypes, tt.cTypes, got, tt.want)
			}
		})
	}
}
