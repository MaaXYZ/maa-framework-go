package checker

import (
	"strings"
	"testing"
)

func TestEvalCConstSources_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		sources []cConstSource
		want    map[string]string
	}{
		{
			name: "integer division truncates toward zero",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `#define MaaFixture_Positive (5 / 2)
#define MaaFixture_Negative (-5 / 2)
#define MaaFixture_Nested ((5 / 2) % 2)
`,
			}},
			want: map[string]string{
				"MaaFixture_Positive": "2",
				"MaaFixture_Negative": "-2",
				"MaaFixture_Nested":   "0",
			},
		},
		{
			name: "macro helper dependencies",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `#define MAA_HELPER_BASE (1ULL << 4)
#define MaaAdbScreencapMethod_Derived (MAA_HELPER_BASE | 1ULL)
#define MaaAdbScreencapMethod_All (~0ULL)
`,
			}},
			want: map[string]string{
				"MaaAdbScreencapMethod_Derived": "17",
				"MaaAdbScreencapMethod_All":     "18446744073709551615",
			},
		},
		{
			name: "enum explicit implicit and signed values",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `enum MaaFixtureEnum
{
    MaaFixture_Zero,
    MaaFixture_Negative = -2,
    MaaFixture_Inherited,
    MaaFixture_Explicit = 4,
    MaaFixture_Dependent = MaaFixture_Explicit + 1,
};
`,
			}},
			want: map[string]string{
				"MaaFixture_Zero":      "0",
				"MaaFixture_Negative":  "-2",
				"MaaFixture_Inherited": "-1",
				"MaaFixture_Explicit":  "4",
				"MaaFixture_Dependent": "5",
			},
		},
		{
			name: "enum entries resolve macros declared later",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `enum MaaFixtureEnum
{
    MaaFixture_First = AFTER,
    MaaFixture_Second,
};

#define AFTER 10
`,
			}},
			want: map[string]string{
				"MaaFixture_First":  "10",
				"MaaFixture_Second": "11",
			},
		},
		{
			name: "integer suffixes character literals and shifts",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `#define MaaFixture_Hex 0x10u
#define MaaFixture_ULL 7ULL
#define MaaFixture_Char 'A'
#define MaaFixture_Shift (1ULL << 40)
#define MaaFixture_Compare (1 > 0)
`,
			}},
			want: map[string]string{
				"MaaFixture_Hex":     "16",
				"MaaFixture_ULL":     "7",
				"MaaFixture_Char":    "65",
				"MaaFixture_Shift":   "1099511627776",
				"MaaFixture_Compare": "1",
			},
		},
		{
			name: "typedef casts preserve signedness and wrap unsigned",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `typedef int32_t MaaFixtureId;
typedef uint32_t MaaFixtureMask;

#define MaaFixture_CastSigned ((MaaFixtureId)(-1))
#define MaaFixture_CastUnsigned ((MaaFixtureMask)(-1))
`,
			}},
			want: map[string]string{
				"MaaFixture_CastSigned":   "-1",
				"MaaFixture_CastUnsigned": "4294967295",
			},
		},
		{
			name: "complements casts and common integer promotions",
			sources: []cConstSource{{
				path: "MaaDef.h",
				content: `typedef uint8_t MaaSmallMask;
#define MAA_NARROW_HELPER 0U
#define MaaFixture_Narrow (~MAA_NARROW_HELPER)
#define MaaFixture_Wide (~0ULL)
#define MaaFixture_Widened ((uint64_t)(~0U))
#define MaaFixture_SmallPromoted (~((MaaSmallMask)(0)))
#define MaaFixture_MixedSigned (-1 < 1ULL)
#define MaaFixture_CommonSigned ((int64_t)(-1) < 1U)
#define MaaFixture_UnsignedWrap (0xffffffffU + 1U)
#define MaaFixture_WideDivide ((~0ULL) / 2ULL)
#define MaaFixture_CastTruncated ((MaaSmallMask)(300))
#define MaaFixture_ShortCircuit (0 && (1 / 0))
`,
			}},
			want: map[string]string{
				"MaaFixture_Narrow":        "4294967295",
				"MaaFixture_Wide":          "18446744073709551615",
				"MaaFixture_Widened":       "4294967295",
				"MaaFixture_SmallPromoted": "-1",
				"MaaFixture_MixedSigned":   "0",
				"MaaFixture_CommonSigned":  "1",
				"MaaFixture_UnsignedWrap":  "0",
				"MaaFixture_WideDivide":    "9223372036854775807",
				"MaaFixture_CastTruncated": "44",
				"MaaFixture_ShortCircuit":  "0",
			},
		},
		{
			name: "helper macro from another configured header",
			sources: []cConstSource{
				{path: "base.h", content: "#define MAA_HELPER_BASE 3\n"},
				{path: "derived.h", content: "#define MaaFixture_Derived (MAA_HELPER_BASE * 2)\n"},
			},
			want: map[string]string{"MaaFixture_Derived": "6"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env, err := evaluateCConstSources(tc.sources)
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			if len(env.failures) != 0 {
				t.Fatalf("unexpected failures: %+v", env.failures)
			}
			for name, want := range tc.want {
				value, ok := env.values[name]
				if !ok {
					t.Fatalf("constant %s not evaluated; failures=%+v", name, env.failures)
				}
				if got := value.ExactString(); got != want {
					t.Errorf("constant %s: got %s want %s", name, got, want)
				}
			}
		})
	}
}

func TestEvalCConstSources_UnsignedWrapsAllOnes(t *testing.T) {
	t.Parallel()

	env, err := evaluateCConstSources([]cConstSource{{
		path:    "MaaDef.h",
		content: "#define MaaFixture_All (~0ULL)\n",
	}})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	value, ok := env.values["MaaFixture_All"]
	if !ok {
		t.Fatalf("constant not evaluated; failures=%+v", env.failures)
	}
	normalized, err := normalizeCNumericValue(numericValue{kind: numericUnsigned, bits: 64, val: value})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got, want := normalized.ExactString(), "18446744073709551615"; got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestEvalCConstSources_DiagnosesFailures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		decl    string
		wantErr string
	}{
		{
			name:    "platform dependent long literal",
			content: "#define Broken (~0UL)\n",
			decl:    "Broken",
			wantErr: "platform-dependent C long literal",
		},
		{
			name:    "signed arithmetic overflow",
			content: "#define Broken (2147483647 + 1)\n",
			decl:    "Broken",
			wantErr: "does not fit signed 32-bit",
		},
		{
			name:    "oversized literal",
			content: "#define Broken 18446744073709551616ULL\n",
			decl:    "Broken",
			wantErr: "outside supported C ranges",
		},
		{
			name:    "oversized unsigned shift",
			content: "#define Broken (1U << 32)\n",
			decl:    "Broken",
			wantErr: "undefined C shift count",
		},
		{
			name:    "negative signed right shift",
			content: "#define Broken (-1 >> 1)\n",
			decl:    "Broken",
			wantErr: "unsupported C shift of negative signed value",
		},
		{
			name:    "unknown identifier",
			content: "#define Broken MISSING\n",
			decl:    "Broken",
			wantErr: "unknown identifier: MISSING",
		},
		{
			name:    "dependency on broken helper",
			content: "#define BrokenBase MISSING\n#define BrokenDerived (BrokenBase + 1)\n",
			decl:    "BrokenDerived",
			wantErr: "unknown identifier: BrokenBase",
		},
		{
			name:    "dependency cycle",
			content: "#define BrokenA BrokenB\n#define BrokenB BrokenA\n",
			decl:    "BrokenA",
			wantErr: "unknown identifier: BrokenB",
		},
		{
			name:    "division by zero",
			content: "#define Broken (1 / 0)\n",
			decl:    "Broken",
			wantErr: "division by zero",
		},
		{
			name:    "modulo by zero",
			content: "#define Broken (1 % 0)\n",
			decl:    "Broken",
			wantErr: "modulo by zero",
		},
		{
			name:    "unsupported function call",
			content: "#define Broken SOME_MACRO(1)\n",
			decl:    "Broken",
			wantErr: "unsupported call expression: SOME_MACRO",
		},
		{
			name:    "unsupported expression shape",
			content: "#define Broken (1 ? 2 : 3)\n",
			decl:    "Broken",
			wantErr: "parse C expression",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			env, err := evaluateCConstSources([]cConstSource{{path: "MaaDef.h", content: tc.content}})
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			failure, ok := env.failures[tc.decl]
			if !ok {
				t.Fatalf("expected failure for %s, got values=%+v failures=%+v", tc.decl, env.values, env.failures)
			}
			if !strings.Contains(failure, tc.wantErr) {
				t.Fatalf("failure for %s: got %q want substring %q", tc.decl, failure, tc.wantErr)
			}
		})
	}
}

func TestEvalCConstSources_RejectsUnparsedEnumMembers(t *testing.T) {
	t.Parallel()
	for _, entry := range []string{"MaaFixture_A [[deprecated]] = 1", "MaaFixture_A =", "MaaFixture_A; MaaFixture_B"} {
		t.Run(entry, func(t *testing.T) {
			t.Parallel()
			_, err := evaluateCConstSources([]cConstSource{{path: "fixture.h", content: "enum Fixture { " + entry + " };"}})
			if err == nil {
				t.Fatalf("expected unsupported member %q to fail parsing", entry)
			}
		})
	}
}
