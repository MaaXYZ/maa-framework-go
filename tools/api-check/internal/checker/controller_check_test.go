package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const controllerTestDefs = `typedef uint8_t MaaBool;
typedef uint64_t MaaControllerFeature;
`

const controllerTestCallbacksHeader = controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(void* trans_arg);
	MaaControllerFeature (*get_features)(void* trans_arg);
};
`

const controllerTestGoInterface = `package maa

type ControllerFeature uint64

type CustomController interface {
	Connect() bool
	GetFeature() ControllerFeature
}
`

// writeControllerTestFixture lays out the header below MaaFramework/Instance
// so parseCTypedefAliases and parseCustomControllerHeader both resolve.
func writeControllerTestFixture(t *testing.T, headerBody, goSource string) (string, string) {
	t.Helper()
	root := t.TempDir()
	headerPath := writeCHeaderFixture(t, root, controllerHeaderRel, headerBody)
	goPath := filepath.Join(root, "custom_controller.go")
	if err := os.WriteFile(goPath, []byte(goSource), 0o600); err != nil {
		t.Fatalf("write %s: %v", goPath, err)
	}
	return headerPath, goPath
}

func writeControllerGoSource(t *testing.T, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "custom_controller.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestParseControllerCallbackFieldWithNestedParens(t *testing.T) {
	t.Parallel()

	stmt := "MaaBool (*request_param)(MaaStringBuffer out, MaaBool (*accept)(int, int (*next)(void)), void* trans_arg)"
	retType, name, paramsRaw, ok := parseControllerCallbackField(stmt)
	if !ok {
		t.Fatalf("expected callback field to parse")
	}
	if retType != "MaaBool" {
		t.Fatalf("unexpected return type: %s", retType)
	}
	if name != "request_param" {
		t.Fatalf("unexpected callback name: %s", name)
	}
	want := "MaaStringBuffer out, MaaBool (*accept)(int, int (*next)(void)), void* trans_arg"
	if paramsRaw != want {
		t.Fatalf("unexpected params: got=%q want=%q", paramsRaw, want)
	}
}

func TestParseControllerCallbackFields(t *testing.T) {
	t.Parallel()

	block := `
	/* callback block */
	MaaBool (*request_uuid)(MaaStringBuffer out, void* trans_arg);
	MaaBool (*do_something)(
		MaaBool (*cb)(int, int (*next)(void)),
		void* trans_arg
	);
`
	fields, err := parseControllerCallbackFields(block)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("unexpected callback count: %d", len(fields))
	}
	if fields[0].name != "request_uuid" {
		t.Fatalf("unexpected first callback: %s", fields[0].name)
	}
	if fields[1].name != "do_something" {
		t.Fatalf("unexpected second callback: %s", fields[1].name)
	}
}

func TestParseControllerCallbackFieldForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		stmt       string
		wantOK     bool
		wantRet    string
		wantName   string
		wantParams string
	}{
		{
			name:       "canonical function pointer",
			stmt:       "MaaBool (*connect)(void* trans_arg)",
			wantOK:     true,
			wantRet:    "MaaBool",
			wantName:   "connect",
			wantParams: "void* trans_arg",
		},
		{
			name:       "pointer with spaces",
			stmt:       "MaaBool ( * connect)(void* trans_arg)",
			wantOK:     true,
			wantRet:    "MaaBool",
			wantName:   "connect",
			wantParams: "void* trans_arg",
		},
		{
			name:       "MAA_CALL function pointer",
			stmt:       "MaaBool (MAA_CALL* connect)(void* trans_arg)",
			wantOK:     true,
			wantRet:    "MaaBool",
			wantName:   "connect",
			wantParams: "void* trans_arg",
		},
		{
			name:   "plain function declaration",
			stmt:   "MaaBool connect(void* trans_arg)",
			wantOK: false,
		},
		{
			name:   "anonymous function pointer",
			stmt:   "MaaBool (*)(void* trans_arg)",
			wantOK: false,
		},
		{
			name:   "invalid callback identifier",
			stmt:   "MaaBool (*1bad)(void* trans_arg)",
			wantOK: false,
		},
		{
			name:   "trailing tokens after parameter list",
			stmt:   "MaaBool (*connect)(void* trans_arg) extra",
			wantOK: false,
		},
		{
			name:   "unbalanced parameter list",
			stmt:   "MaaBool (*connect(void* trans_arg)",
			wantOK: false,
		},
		{
			name:   "empty statement",
			stmt:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ret, name, params, ok := parseControllerCallbackField(tt.stmt)
			if ok != tt.wantOK {
				t.Fatalf("parseControllerCallbackField(%q) ok = %v, want %v", tt.stmt, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if ret != tt.wantRet || name != tt.wantName || params != tt.wantParams {
				t.Fatalf("parseControllerCallbackField(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tt.stmt, ret, name, params, tt.wantRet, tt.wantName, tt.wantParams)
			}
		})
	}
}

func TestParseControllerCallbackFieldsRejectsUnsupported(t *testing.T) {
	t.Parallel()

	t.Run("unsupported field among valid fields", func(t *testing.T) {
		t.Parallel()
		block := `
MaaBool (*connect)(void* trans_arg);
MaaBool broken(void* trans_arg);
MaaBool (*connected)(void* trans_arg);
`
		if _, err := parseControllerCallbackFields(block); err == nil || !strings.Contains(err.Error(), "unsupported controller callback field") {
			t.Fatalf("parseControllerCallbackFields() error = %v, want unsupported field error", err)
		}
	})

	t.Run("spaced pointer and comments accepted", func(t *testing.T) {
		t.Parallel()
		block := `
/* leading comment */
MaaBool ( * connect)(void* trans_arg); // trailing comment
MaaBool (*request_uuid)(void* trans_arg, /* out */ MaaStringBuffer* buffer);
`
		fields, err := parseControllerCallbackFields(block)
		if err != nil {
			t.Fatalf("parseControllerCallbackFields() error = %v", err)
		}
		if len(fields) != 2 {
			t.Fatalf("callback count = %d, want 2", len(fields))
		}
		if fields[0].name != "connect" || fields[1].name != "request_uuid" {
			t.Fatalf("unexpected callbacks: %+v", fields)
		}
		if got := normalizeSpaces(fields[1].paramsRaw); got != "void* trans_arg, MaaStringBuffer* buffer" {
			t.Fatalf("request_uuid params = %q", got)
		}
	})
}

func TestControllerConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		goSource   string
		wantIssues []string
		denyIssues []string
	}{
		{
			name:     "matching interface and callbacks",
			header:   controllerTestCallbacksHeader,
			goSource: controllerTestGoInterface,
		},
		{
			name:     "C callback missing in Go interface",
			header:   controllerTestCallbacksHeader,
			goSource: "package maa\n\ntype CustomController interface {\n\tConnect() bool\n}\n",
			wantIssues: []string{
				"C callback missing in Go interface: GetFeature",
			},
		},
		{
			name:   "Go interface method missing in C callbacks",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

import "image"

type ControllerFeature uint64

type CustomController interface {
	Connect() bool
	GetFeature() ControllerFeature
	Screencap() (image.Image, bool)
}
`,
			wantIssues: []string{
				"Go interface method not found in C callbacks: Screencap",
			},
		},
		{
			name:   "parameter mismatch",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type ControllerFeature uint64

type CustomController interface {
	Connect(x int32) bool
	GetFeature() ControllerFeature
}
`,
			wantIssues: []string{
				"Connect param mismatch: go=[int32] c=[]",
			},
		},
		{
			name:   "return mismatch",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type ControllerFeature uint64

type CustomController interface {
	Connect() int32
	GetFeature() ControllerFeature
}
`,
			wantIssues: []string{
				"Connect return mismatch: go=[int32] c=[bool]",
			},
		},
		{
			name:   "unsupported Go parameter expression",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type ControllerFeature uint64

type CustomController interface {
	Connect(handle []byte) bool
	GetFeature() ControllerFeature
}
`,
			wantIssues: []string{
				"Connect has unsupported Go param type expression",
			},
			denyIssues: []string{"Connect param mismatch"},
		},
		{
			name:   "unsupported Go return expression",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type ControllerFeature uint64

type CustomController interface {
	Connect() []byte
	GetFeature() ControllerFeature
}
`,
			wantIssues: []string{
				"Connect has unsupported Go return type expression",
			},
			denyIssues: []string{"Connect return mismatch"},
		},
		{
			name: "spaced callback pointer is accepted",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool ( * connect)(void* trans_arg);
	MaaControllerFeature ( * get_features)(void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
		},
		{
			name: "output adapters map to Go values",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(void* trans_arg, /* out */ MaaStringBuffer* buffer);
	MaaBool (*screencap)(void* trans_arg, /* out */ MaaImageBuffer* buffer);
	MaaBool (*shell)(const char* cmd, int64_t timeout, void* trans_arg, /* out */ MaaStringBuffer* buffer);
	MaaBool (*get_info)(void* trans_arg, /* out */ MaaStringBuffer* buffer);
};
`,
			goSource: `package maa

import "image"

type CustomController interface {
	RequestUUID() (string, bool)
	Screencap() (image.Image, bool)
	Shell(cmd string, timeout int64) (string, bool)
	GetInfo() (string, bool)
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerPath, goPath := writeControllerTestFixture(t, tt.header, tt.goSource)
			issues, err := checkCustomControllerConsistency(headerPath, goPath)
			if err != nil {
				t.Fatalf("checkCustomControllerConsistency() error = %v", err)
			}
			for _, want := range tt.wantIssues {
				assertIssueContains(t, issues, want)
			}
			for _, deny := range tt.denyIssues {
				assertIssueNotContains(t, issues, deny)
			}
			if len(tt.wantIssues) == 0 && len(issues) != 0 {
				t.Fatalf("expected no issues, got: %+v", issues)
			}
		})
	}
}

func TestControllerConsistencyErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		header   string
		goSource string
		wantErr  string
	}{
		{
			name: "unsupported C callback field",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool connect(void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported controller callback field",
		},
		{
			name: "output buffer before context",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(MaaStringBuffer* buffer, void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "context parameter is not last",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*click)(void* trans_arg, int32_t x);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported trans_arg ABI (expected void* at parameter 2)",
		},
		{
			name: "context parameter is not void pointer",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(char* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported trans_arg ABI",
		},
		{
			name: "missing context parameter",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(int32_t x);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "expected exactly one trans_arg",
		},
		{
			name: "duplicate context parameters",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(void* trans_arg, void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported trans_arg ABI",
		},
		{
			name: "wrong output buffer type for adapter",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*screencap)(void* trans_arg, MaaStringBuffer* buffer);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "multiple output buffers cannot collapse into one return",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(MaaStringBuffer* first, void* trans_arg, MaaStringBuffer* second);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "ordinary callback cannot acquire an output adapter",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*connect)(MaaStringBuffer* buffer, void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "double pointer output buffer",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(void* trans_arg, MaaStringBuffer** buffer);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "output buffer passed by value",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(void* trans_arg, MaaStringBuffer buffer);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "unsupported output buffer ABI",
		},
		{
			name: "missing output buffer for adapter",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(void* trans_arg, int32_t x);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "expected 1 output buffer, got 0",
		},
		{
			name: "output adapter requires MaaBool return",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaControllerFeature (*request_uuid)(void* trans_arg, MaaStringBuffer* buffer);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "output adapter requires MaaBool return",
		},
		{
			name: "duplicate controller callback",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
	MaaBool (*request_uuid)(void* trans_arg, MaaStringBuffer* buffer);
	MaaBool (*request_uuid)(void* trans_arg, MaaStringBuffer* buffer);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "duplicate controller callback RequestUUID",
		},
		{
			name: "callbacks struct missing",
			header: controllerTestDefs + `
struct MaaOtherCallbacks {
	MaaBool (*connect)(void* trans_arg);
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "MaaCustomControllerCallbacks struct not found",
		},
		{
			name: "callbacks struct is empty",
			header: controllerTestDefs + `
struct MaaCustomControllerCallbacks {
};
`,
			goSource: controllerTestGoInterface,
			wantErr:  "no callback field found in callbacks struct",
		},
		{
			name:     "Go interface missing",
			header:   controllerTestCallbacksHeader,
			goSource: "package maa\n",
			wantErr:  "CustomController interface not found",
		},
		{
			name:   "Go embedded interface cycle",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type CustomController interface {
	Other
}

type Other interface {
	CustomController
}
`,
			wantErr: "embedded interface cycle",
		},
		{
			name:   "Go unsupported embedded selector",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type CustomController interface {
	other.Base
}
`,
			wantErr: "unsupported embedded interface",
		},
		{
			name:   "Go CustomController is not an interface",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type CustomController struct{}
`,
			wantErr: "unresolved or unsupported embedded interface CustomController",
		},
		{
			name:   "Go unsupported type term element",
			header: controllerTestCallbacksHeader,
			goSource: `package maa

type CustomController interface {
	~int32 | ~string
}
`,
			wantErr: "unsupported embedded interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headerPath, goPath := writeControllerTestFixture(t, tt.header, tt.goSource)
			_, err := checkCustomControllerConsistency(headerPath, goPath)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("checkCustomControllerConsistency() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseCustomControllerGoEmbedding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     string
		want    map[string][2]string
		wantErr string
	}{
		{
			name: "grouped parameters and results",
			src: `package maa

type CustomController interface {
	Connect(handle uintptr, name string) bool
	Click(x, y int32) (a, b bool)
}
`,
			want: map[string][2]string{
				"Connect": {"ptr,cstring", "bool"},
				"Click":   {"int32,int32", "bool,bool"},
			},
		},
		{
			name: "local embedded interfaces",
			src: `package maa

type Core interface {
	Connect(handle uintptr) bool
}

type Base interface {
	Core
	GetFeature() uint64
}

type CustomController interface {
	Base
	Click(x, y int32) bool
}
`,
			want: map[string][2]string{
				"Connect":    {"ptr", "bool"},
				"GetFeature": {"", "uint64"},
				"Click":      {"int32,int32", "bool"},
			},
		},
		{
			name: "identical methods from two embedded interfaces",
			src: `package maa

type A interface{ Connect() bool }
type B interface{ Connect() bool }

type CustomController interface {
	A
	B
}
`,
			want: map[string][2]string{"Connect": {"", "bool"}},
		},
		{
			name: "conflicting methods from two embedded interfaces",
			src: `package maa

type A interface{ Connect() bool }
type B interface{ Connect() int32 }

type CustomController interface {
	A
	B
}
`,
			wantErr: "conflicting embedded method Connect",
		},
		{
			name: "self cycle",
			src: `package maa

type CustomController interface {
	CustomController
}
`,
			wantErr: "cycle",
		},
		{
			name: "indirect cycle",
			src: `package maa

type A interface{ B }
type B interface{ A }

type CustomController interface {
	A
}
`,
			wantErr: "cycle",
		},
		{
			name: "type alias to an interface",
			src: `package maa

type Base interface{ Connect() bool }

type CustomController = Base
`,
			want: map[string][2]string{"Connect": {"", "bool"}},
		},
		{
			name: "embedded struct is not an interface",
			src: `package maa

type Base struct{}

type CustomController interface {
	Base
}
`,
			wantErr: "unresolved or unsupported embedded interface Base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sigs, err := parseCustomControllerGo(writeControllerGoSource(t, tt.src))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseCustomControllerGo() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCustomControllerGo() error = %v", err)
			}
			for name, want := range tt.want {
				sig, ok := sigs[name]
				if !ok {
					t.Fatalf("missing method %s (have %v)", name, sigs)
				}
				if got := strings.Join(sig.params, ","); got != want[0] {
					t.Errorf("%s params = %q, want %q", name, got, want[0])
				}
				if got := strings.Join(sig.returns, ","); got != want[1] {
					t.Errorf("%s returns = %q, want %q", name, got, want[1])
				}
			}
		})
	}
}

func TestControllerCallbackNameToGoMethod(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"request_uuid": "RequestUUID",
		"get_features": "GetFeature",
		"touch_down":   "TouchDown",
		"click":        "Click",
		"get_info":     "GetInfo",
	}
	for cName, goName := range tests {
		if got := callbackNameToGoMethod(cName); got != goName {
			t.Errorf("callbackNameToGoMethod(%q) = %q, want %q", cName, got, goName)
		}
	}
}
