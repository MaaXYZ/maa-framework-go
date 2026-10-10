package checker

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCDefineDecls_PreservesDeclarations(t *testing.T) {
	t.Parallel()

	content := "#pragma once\r\n" +
		"#define MAA_BASE 1ULL\r\n" +
		"#define MAA_FUNCTION(value) ((value) + 1)\r\n" +
		"#define MAA_MULTILINE (MAA_BASE | \\\r\n" +
		"    2ULL)\r\n" +
		"#if MAA_UNKNOWN\r\n" +
		"#define MAA_BASE 1ULL\r\n" +
		"#else\r\n" +
		"#define MAA_BASE 2ULL\r\n" +
		"#endif\r\n" +
		"#define MAA_EMPTY\r\n"
	want := []cConstDecl{
		{name: "MAA_BASE", expr: "1ULL"},
		{name: "MAA_MULTILINE", expr: "(MAA_BASE | 2ULL)"},
		{name: "MAA_BASE", expr: "1ULL"},
		{name: "MAA_BASE", expr: "2ULL"},
	}
	if got := parseCDefineDecls(content); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCDefineDecls() = %+v, want %+v", got, want)
	}
}

func TestParseCDefineExprs_PreservesMapBehavior(t *testing.T) {
	t.Parallel()

	content := "#define MaaMsg_Fixture \"first\"\n" +
		"#define MaaMsg_Fixture \"second\"\n" +
		"#define MAA_BASE 1ULL\n"
	want := map[string]string{
		"MaaMsg_Fixture": `"second"`,
		"MAA_BASE":       "1ULL",
	}
	if got := parseCDefineExprs(content); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCDefineExprs() = %+v, want %+v", got, want)
	}
}

func TestEvalCConstSources_MacroDeclarations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		want    map[string]string
		wantErr string
	}{
		{
			name: "unknown guard with conflicting checked macro",
			content: `#if MAA_UNKNOWN
#define MaaMacOSScreencapMethod_ScreenCaptureKit 1ULL
#else
#define MaaMacOSScreencapMethod_ScreenCaptureKit 2ULL
#endif
`,
			wantErr: "conflicting C constant MaaMacOSScreencapMethod_ScreenCaptureKit",
		},
		{
			name: "unknown guard with conflicting helper macro",
			content: `#ifdef MAA_UNKNOWN
#define MAA_HELPER 1ULL
#else
#define MAA_HELPER 2ULL
#endif
enum Fixture { MaaFixture_FromHelper = MAA_HELPER };
`,
			wantErr: "conflicting C constant MAA_HELPER",
		},
		{
			name: "unknown guard with equivalent definitions",
			content: `#if MAA_UNKNOWN
#define MAA_HELPER (1ULL << 2)
#else
#define MAA_HELPER (1ULL    << 2)
#endif
enum Fixture { MaaFixture_FromHelper = MAA_HELPER + 1 };
`,
			want: map[string]string{
				"MAA_HELPER":            "4",
				"MaaFixture_FromHelper": "5",
			},
		},
		{
			name: "repeated equivalent definitions and dependencies",
			content: `#define MAA_BASE (1ULL << 2)
#define MAA_BASE (1ULL    << 2)
#define MAA_DERIVED (MAA_BASE | 8ULL)
enum Fixture { MaaFixture_FromHelper = MAA_DERIVED + 1 };
#define MaaFixture_FromEnum (MaaFixture_FromHelper + 1)
`,
			want: map[string]string{
				"MAA_BASE":              "4",
				"MAA_DERIVED":           "12",
				"MaaFixture_FromHelper": "13",
				"MaaFixture_FromEnum":   "14",
			},
		},
		{
			name: "known C++ guard selects only C macro",
			content: `#if defined(__cplusplus)
#define MaaFixture_Value 99
#else
#define MaaFixture_Value 2
#endif
`,
			want: map[string]string{"MaaFixture_Value": "2"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env, err := evaluateCConstSources([]cConstSource{{path: "fixture.h", content: tc.content}})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("evaluateCConstSources() error = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("evaluateCConstSources() error = %v", err)
			}
			if len(env.failures) != 0 {
				t.Fatalf("unexpected evaluation failures: %+v", env.failures)
			}
			for name, want := range tc.want {
				value, ok := env.values[name]
				if !ok {
					t.Errorf("missing constant %s", name)
				} else if got := value.ExactString(); got != want {
					t.Errorf("%s = %s, want %s", name, got, want)
				}
			}
		})
	}
}
