package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallbackABIMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		goSig methodSig
		cSig  methodSig
		want  bool
	}{
		{
			name:  "identical signatures",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"uint64"}},
			cSig:  methodSig{params: []string{"ptr"}, returns: []string{"uint64"}},
			want:  true,
		},
		{
			name:  "C void still requires a pointer-sized Go result",
			goSig: methodSig{},
			cSig:  methodSig{},
			want:  false,
		},
		{
			name:  "C strings accept byte pointers",
			goSig: methodSig{params: []string{"ptr", "ptr", "ptr", "ptr"}, returns: []string{"ptr"}},
			cSig:  methodSig{params: []string{"ptr", "cstring", "cstring", "ptr"}},
			want:  true,
		},
		{
			name:  "Go strings are unsupported by purego callbacks",
			goSig: methodSig{params: []string{"cstring"}},
			cSig:  methodSig{params: []string{"cstring"}},
			want:  false,
		},
		{
			name:  "Go string must not replace a C pointer",
			goSig: methodSig{params: []string{"cstring"}},
			cSig:  methodSig{params: []string{"ptr"}},
			want:  false,
		},
		{
			name:  "parameter arity mismatch",
			goSig: methodSig{params: []string{"ptr"}},
			cSig:  methodSig{params: []string{"ptr", "ptr"}},
			want:  false,
		},
		{
			name:  "parameter order mismatch",
			goSig: methodSig{params: []string{"ptr", "cstring"}},
			cSig:  methodSig{params: []string{"cstring", "ptr"}},
			want:  false,
		},
		{
			name:  "uintptr return matches C void",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"ptr"}},
			cSig:  methodSig{params: []string{"ptr"}},
			want:  true,
		},
		{
			name:  "uintptr return matches C MaaBool",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"ptr"}},
			cSig:  methodSig{params: []string{"ptr"}, returns: []string{"bool"}},
			want:  true,
		},
		{
			name:  "uintptr return does not match other C returns",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"ptr"}},
			cSig:  methodSig{params: []string{"ptr"}, returns: []string{"int32"}},
			want:  false,
		},
		{
			name:  "non-pointer return does not match C void",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"int32"}},
			cSig:  methodSig{params: []string{"ptr"}},
			want:  false,
		},
		{
			name:  "byte pointer return matches C string",
			goSig: methodSig{returns: []string{"ptr"}},
			cSig:  methodSig{returns: []string{"cstring"}},
			want:  true,
		},
		{
			name:  "Go string return is unsupported",
			goSig: methodSig{returns: []string{"cstring"}},
			cSig:  methodSig{returns: []string{"cstring"}},
			want:  false,
		},
		{
			name:  "matching bool returns",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"bool"}},
			cSig:  methodSig{params: []string{"ptr"}, returns: []string{"bool"}},
			want:  false,
		},
		{
			name:  "matching small integer returns",
			goSig: methodSig{returns: []string{"int32"}},
			cSig:  methodSig{returns: []string{"int32"}},
			want:  false,
		},
		{
			name:  "matching floating-point returns",
			goSig: methodSig{returns: []string{"float64"}},
			cSig:  methodSig{returns: []string{"float64"}},
			want:  false,
		},
		{
			name:  "matching signed pointer-sized returns on supported targets",
			goSig: methodSig{returns: []string{"int64"}},
			cSig:  methodSig{returns: []string{"int64"}},
			want:  true,
		},
		{
			name:  "return arity mismatch",
			goSig: methodSig{params: []string{"ptr"}, returns: []string{"ptr", "bool"}},
			cSig:  methodSig{params: []string{"ptr"}, returns: []string{"bool"}},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := callbackABIMatches(tt.goSig, tt.cSig); got != tt.want {
				t.Fatalf("callbackABIMatches(%v -> %v, %v -> %v) = %v, want %v",
					tt.goSig.params, tt.goSig.returns, tt.cSig.params, tt.cSig.returns, got, tt.want)
			}
		})
	}
}

func TestParseCNamedCallbacksFromDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeCHeaderFixture(t, dir, "MaaFramework/MaaDef.h", `typedef uint8_t MaaBool;

typedef void(MAA_CALL* MaaEventCallback)(void* handle, const char* message, const char* details_json, void* trans_arg);

typedef void (*MaaSimpleCallback)(int32_t code);

typedef MaaBool(MAA_CALL* MaaCustomActionCallback)(
    void* context,
    int64_t task_id,
    const char* current_task_name
);

/* No Callback suffix, so it is not collected. */
typedef void (*MaaNotification)(const char* message);
`)

	callbacks, err := parseCNamedCallbacks(dir, nil)
	if err != nil {
		t.Fatalf("parseCNamedCallbacks() error = %v", err)
	}

	assertCallback := func(name, params, returns string) {
		t.Helper()
		sig, ok := callbacks[name]
		if !ok {
			t.Fatalf("callback %s not found (have %v)", name, callbacks)
		}
		if got := strings.Join(sig.params, ","); got != params {
			t.Errorf("%s params = %q, want %q", name, got, params)
		}
		if got := strings.Join(sig.returns, ","); got != returns {
			t.Errorf("%s returns = %q, want %q", name, got, returns)
		}
	}
	assertCallback("MaaEventCallback", "ptr,cstring,cstring,ptr", "")
	assertCallback("MaaSimpleCallback", "int32", "")
	assertCallback("MaaCustomActionCallback", "ptr,int64,cstring", "bool")
	if _, ok := callbacks["MaaNotification"]; ok {
		t.Errorf("typedef without Callback suffix must not be collected")
	}
}

func TestParseCNamedCallbacksErrors(t *testing.T) {
	t.Parallel()

	t.Run("unsupported callback typedef", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeCHeaderFixture(t, dir, "MaaDef.h", "typedef void MaaBadCallback(void* trans_arg);\n")
		if _, err := parseCNamedCallbacks(dir, nil); err == nil || !strings.Contains(err.Error(), "unsupported callback typedef") {
			t.Fatalf("parseCNamedCallbacks() error = %v, want unsupported typedef error", err)
		}
	})

	t.Run("conflicting callback typedefs", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeCHeaderFixture(t, dir, "a.h", "typedef void (*MaaSimpleCallback)(int32_t code);\n")
		writeCHeaderFixture(t, dir, "b.h", "typedef void (*MaaSimpleCallback)(int32_t code, void* extra);\n")
		if _, err := parseCNamedCallbacks(dir, nil); err == nil || !strings.Contains(err.Error(), "conflicting C callback typedef MaaSimpleCallback") {
			t.Fatalf("parseCNamedCallbacks() error = %v, want conflict error", err)
		}
	})

	t.Run("identical callback typedefs agree", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		body := "typedef void (*MaaSimpleCallback)(int32_t code);\n"
		writeCHeaderFixture(t, dir, "a.h", body)
		writeCHeaderFixture(t, dir, "b.h", body)
		callbacks, err := parseCNamedCallbacks(dir, nil)
		if err != nil {
			t.Fatalf("parseCNamedCallbacks() error = %v", err)
		}
		if got := strings.Join(callbacks["MaaSimpleCallback"].params, ","); got != "int32" {
			t.Fatalf("MaaSimpleCallback params = %q, want int32", got)
		}
	})
}

const callbackLayoutHeaderTwo = `typedef uint8_t MaaBool;

struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(void* trans_arg);
	MaaBool (*get_info)(void* trans_arg, MaaStringBuffer* buffer);
};
`

const callbackLayoutGoTwo = `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`

func writeCallbackLayoutFixture(t *testing.T, header, goSource string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	headerPath := writeCHeaderFixture(t, dir, "MaaCustomController.h", header)
	goPath := filepath.Join(dir, "custom_controller.go")
	if err := os.WriteFile(goPath, []byte(goSource), 0o600); err != nil {
		t.Fatalf("write %s: %v", goPath, err)
	}
	return headerPath, goPath
}

func TestCheckControllerCallbackLayout(t *testing.T) {
	t.Parallel()

	t.Run("matching struct order and bindings", func(t *testing.T) {
		t.Parallel()
		headerPath, goPath := writeCallbackLayoutFixture(t, callbackLayoutHeaderTwo, callbackLayoutGoTwo)
		issues, err := checkControllerCallbackLayout(headerPath, goPath, nil)
		if err != nil {
			t.Fatalf("checkControllerCallbackLayout() error = %v", err)
		}
		if len(issues) != 0 {
			t.Fatalf("expected no issues, got: %+v", issues)
		}
	})

	t.Run("specialized callback target name matches C field", func(t *testing.T) {
		t.Parallel()
		header := `typedef uint8_t MaaBool;

struct MaaCustomControllerCallbacks {
	MaaBool (*click_key)(int32_t keycode, void* trans_arg);
};
`
		goSource := `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	ClickKey uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.ClickKey = purego.NewCallback(_ClickKey)
}

func _ClickKey(keycode int32, transArg uintptr) uintptr { return 0 }
`
		headerPath, goPath := writeCallbackLayoutFixture(t, header, goSource)
		issues, err := checkControllerCallbackLayout(headerPath, goPath, nil)
		if err != nil {
			t.Fatalf("checkControllerCallbackLayout() error = %v", err)
		}
		if len(issues) != 0 {
			t.Fatalf("expected no issues, got: %+v", issues)
		}
	})

	tests := []struct {
		name       string
		header     string
		goSource   string
		wantIssues []string
	}{
		{
			name:   "binding target name does not match callback field",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_GetInfoAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback binding target mismatch for Connect: got _GetInfoAgent, want _ConnectAgent"},
		},
		{
			name:   "binding to another function with identical ABI is rejected",
			header: callbackLayoutHeaderTwo,
			goSource: strings.Replace(callbackLayoutGoTwo, "NewCallback(_ConnectAgent)", "NewCallback(_ConnectedAgent)", 1) + `
func _ConnectedAgent(transArg uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback binding target mismatch for Connect: got _ConnectedAgent, want _ConnectAgent"},
		},
		{
			name: "specialized binding target name mismatch",
			header: `typedef uint8_t MaaBool;

struct MaaCustomControllerCallbacks {
	MaaBool (*click_key)(int32_t keycode, void* trans_arg);
};
`,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	ClickKey uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.ClickKey = purego.NewCallback(_ClickKeyAgent)
}

func _ClickKeyAgent(keycode int32, transArg uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback binding target mismatch for ClickKey: got _ClickKeyAgent, want _ClickKey"},
		},
		{
			name:   "struct field order mismatch",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	GetInfo uintptr
	Connect uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback struct layout mismatch: go=[GetInfo Connect] c=[Connect GetInfo]"},
		},
		{
			name:   "struct field deleted",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback struct layout mismatch: go=[Connect] c=[Connect GetInfo]"},
		},
		{
			name:   "missing binding for callback field",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"missing callback binding or target for GetInfo"},
		},
		{
			name:   "binding target function missing",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_MissingAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"missing callback binding or target for Connect"},
		},
		{
			name:   "binding target ABI mismatch",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
}

func _ConnectAgent(transArg uintptr) int32 { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }
`,
			wantIssues: []string{"callback binding ABI mismatch for Connect"},
		},
		{
			name:   "binding absent from C struct",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
	Extra   uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.GetInfo = purego.NewCallback(_GetInfoAgent)
	customControllerCallbacksHandle.Extra = purego.NewCallback(_ExtraAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }

func _GetInfoAgent(transArg, buffer uintptr) uintptr { return 0 }

func _ExtraAgent(transArg uintptr) uintptr { return 0 }
`,
			wantIssues: []string{
				"Go callback binding absent from C: Extra",
				"callback struct layout mismatch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerPath, goPath := writeCallbackLayoutFixture(t, tt.header, tt.goSource)
			issues, err := checkControllerCallbackLayout(headerPath, goPath, nil)
			if err != nil {
				t.Fatalf("checkControllerCallbackLayout() error = %v", err)
			}
			for _, want := range tt.wantIssues {
				assertIssueContains(t, issues, want)
			}
		})
	}
}

func TestCheckControllerCallbackLayoutErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   string
		goSource string
		wantErr  string
	}{
		{
			name:   "struct field is not pointer sized",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

type MaaCustomControllerCallbacks struct {
	Connect int32
	GetInfo uintptr
}
`,
			wantErr: "callbacks field must have pointer-sized type",
		},
		{
			name:   "struct has an embedded field",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

type MaaCustomControllerCallbacks struct {
	uintptr
	GetInfo uintptr
}
`,
			wantErr: "unsupported embedded callbacks layout",
		},
		{
			name:   "binding does not use purego.NewCallback",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = _ConnectAgent
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`,
			wantErr: "unsupported callback binding",
		},
		{
			name:   "binding uses another callback constructor package",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = other.NewCallback(_ConnectAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`,
			wantErr: "unsupported callback binding package",
		},
		{
			name:   "duplicate callback binding",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
	GetInfo uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`,
			wantErr: "duplicate callback binding Connect",
		},
		{
			name:   "Go callbacks struct missing",
			header: callbackLayoutHeaderTwo,
			goSource: `package maa

import "github.com/ebitengine/purego"

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`,
			wantErr: "MaaCustomControllerCallbacks struct not found",
		},
		{
			name:     "C callbacks struct missing",
			header:   "typedef uint8_t MaaBool;\n",
			goSource: callbackLayoutGoTwo,
			wantErr:  "missing callbacks struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerPath, goPath := writeCallbackLayoutFixture(t, tt.header, tt.goSource)
			_, err := checkControllerCallbackLayout(headerPath, goPath, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("checkControllerCallbackLayout() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

const callbackCoverageDefHeader = `typedef uint8_t MaaBool;

typedef void(MAA_CALL* MaaEventCallback)(void* handle, const char* message, const char* details_json, void* trans_arg);
`

const callbackCoverageControllerHeader = `typedef uint8_t MaaBool;

struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(void* trans_arg);
};
`

const callbackCoverageControllerGo = `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
}

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`

const callbackCoverageEventGo = `package maa

func _MaaEventCallbackAgent(handle uintptr, message, detailsJson *byte, transArg uintptr) uintptr {
	return 0
}
`

const callbackCoverageFrameworkGo = `package native

type MaaEventCallback func(handle uintptr, message, detailsJson *byte, transArg uintptr) uintptr
`

type callbackCoverageFixture struct {
	defHeader        string
	controllerHeader string
	controllerGo     string
	eventGo          string
	frameworkGo      string
}

func writeCallbackCoverageFixture(t *testing.T, fx callbackCoverageFixture) (string, string, map[string][]string) {
	t.Helper()
	root := t.TempDir()
	headerDir := filepath.Join(root, "deps", "include")
	writeCHeaderFixture(t, headerDir, "MaaFramework/MaaDef.h", fx.defHeader)
	writeCHeaderFixture(t, headerDir, filepath.FromSlash(controllerHeaderRel), fx.controllerHeader)

	writeCallbackCoverageSource(t, filepath.Join(root, "custom_controller.go"), fx.controllerGo)
	writeCallbackCoverageSource(t, filepath.Join(root, "event.go"), fx.eventGo)
	frameworkPath := filepath.Join(root, "internal", "native", "framework.go")
	writeCallbackCoverageSource(t, frameworkPath, fx.frameworkGo)
	return root, headerDir, map[string][]string{"framework": {frameworkPath}}
}

func writeCallbackCoverageSource(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCheckCallbackABICoverage(t *testing.T) {
	t.Parallel()

	base := callbackCoverageFixture{
		defHeader:        callbackCoverageDefHeader,
		controllerHeader: callbackCoverageControllerHeader,
		controllerGo:     callbackCoverageControllerGo,
		eventGo:          callbackCoverageEventGo,
		frameworkGo:      callbackCoverageFrameworkGo,
	}

	t.Run("matching named callback and trampoline", func(t *testing.T) {
		t.Parallel()
		root, headerDir, nativeFiles := writeCallbackCoverageFixture(t, base)
		issues, err := checkCallbackABICoverage(root, headerDir, nativeFiles)
		if err != nil {
			t.Fatalf("checkCallbackABICoverage() error = %v", err)
		}
		if len(issues) != 0 {
			t.Fatalf("expected no issues, got: %+v", issues)
		}
	})

	tests := []struct {
		name       string
		mutate     func(*callbackCoverageFixture)
		wantIssues []string
		denyIssues []string
	}{
		{
			name: "missing trampoline",
			mutate: func(fx *callbackCoverageFixture) {
				fx.eventGo = "package maa\n"
			},
			wantIssues: []string{"missing callback trampoline for MaaEventCallback"},
		},
		{
			name: "trampoline ABI mismatch",
			mutate: func(fx *callbackCoverageFixture) {
				fx.eventGo = `package maa

func _MaaEventCallbackAgent(handle uintptr, message, detailsJson *byte, transArg uintptr) int32 {
	return 0
}
`
			},
			wantIssues: []string{"callback trampoline ABI mismatch for MaaEventCallback"},
		},
		{
			name: "matching void typedef and trampoline cannot register on Windows",
			mutate: func(fx *callbackCoverageFixture) {
				fx.frameworkGo = strings.Replace(callbackCoverageFrameworkGo, ") uintptr", ")", 1)
				fx.eventGo = strings.Replace(callbackCoverageEventGo, ") uintptr {\n\treturn 0", ") {", 1)
			},
			wantIssues: []string{"callback ABI mismatch for MaaEventCallback", "callback trampoline ABI mismatch for MaaEventCallback"},
		},
		{
			name: "bool controller trampoline cannot register on Windows",
			mutate: func(fx *callbackCoverageFixture) {
				fx.controllerGo = strings.Replace(callbackCoverageControllerGo,
					"func _ConnectAgent(transArg uintptr) uintptr { return 0 }",
					"func _ConnectAgent(transArg uintptr) bool { return false }", 1)
			},
			wantIssues: []string{"callback binding ABI mismatch for Connect"},
		},
		{
			name: "Go callback typedef absent from C",
			mutate: func(fx *callbackCoverageFixture) {
				fx.defHeader = "typedef uint8_t MaaBool;\n"
			},
			wantIssues: []string{"Go callback typedef absent from C: MaaEventCallback"},
		},
		{
			name: "exported API callback typedef missing in Go",
			mutate: func(fx *callbackCoverageFixture) {
				fx.defHeader = callbackCoverageDefHeader + "\ntypedef void (*MaaSimpleCallback)(int32_t code);\nMAA_FRAMEWORK_API void MaaRegisterSimple(MaaSimpleCallback callback);\n"
			},
			wantIssues: []string{"C callback typedef missing in Go: MaaSimpleCallback"},
		},
		{
			name: "unused retired C callback typedef does not require a Go wrapper",
			mutate: func(fx *callbackCoverageFixture) {
				fx.defHeader = callbackCoverageDefHeader + "\ntypedef void (*MaaNotificationCallback)(int32_t code);\n"
			},
			denyIssues: []string{"C callback typedef missing in Go"},
		},
		{
			name: "named callback ABI mismatch",
			mutate: func(fx *callbackCoverageFixture) {
				fx.frameworkGo = `package native

type MaaEventCallback func(handle uintptr) uintptr
`
			},
			wantIssues: []string{"callback ABI mismatch for MaaEventCallback"},
			denyIssues: []string{"callback trampoline ABI mismatch"},
		},
		{
			name: "controller layout issues are included",
			mutate: func(fx *callbackCoverageFixture) {
				fx.controllerGo = `package maa

import "github.com/ebitengine/purego"

type MaaCustomControllerCallbacks struct {
	Connect uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func _ConnectAgent(transArg uintptr) uintptr { return 0 }
`
			},
			wantIssues: []string{"missing callback binding or target for Connect"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := base
			tt.mutate(&fx)
			root, headerDir, nativeFiles := writeCallbackCoverageFixture(t, fx)
			issues, err := checkCallbackABICoverage(root, headerDir, nativeFiles)
			if err != nil {
				t.Fatalf("checkCallbackABICoverage() error = %v", err)
			}
			for _, want := range tt.wantIssues {
				assertIssueContains(t, issues, want)
			}
			for _, deny := range tt.denyIssues {
				assertIssueNotContains(t, issues, deny)
			}
		})
	}
}
