package checker

import (
	"strings"
	"testing"
)

func TestEvalGoConstSource_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  string
		want map[string]string
	}{
		{
			name: "narrow conversion and complement",
			src: `package fixture

type ScreencapMethod uint64

const (
	Wide     uint64          = uint64(^uint32(0))
	Narrow   ScreencapMethod = ScreencapMethod(^uint32(0))
	ByteWide                 = ^uint8(0)
	All      ScreencapMethod = ^ScreencapMethod(0)
	Zero                     = ^uint32(0) & 0
)
`,
			want: map[string]string{
				"Wide":     "4294967295",
				"Narrow":   "4294967295",
				"ByteWide": "255",
				"All":      "18446744073709551615",
				"Zero":     "0",
			},
		},
		{
			name: "iota and inherited expressions",
			src: `package fixture

type InputMethod uint64

const (
	InputNone InputMethod = iota
	InputOne
	InputShift = 1 << iota
	InputShiftNext
	InputReset InputMethod = iota
)
`,
			want: map[string]string{
				"InputNone":      "0",
				"InputOne":       "1",
				"InputShift":     "4",
				"InputShiftNext": "8",
				"InputReset":     "4",
			},
		},
		{
			name: "shifts negatives dependencies and aliases",
			src: `package fixture

type ScreencapMethod uint64
type Signed int32

const (
	helperBase       ScreencapMethod = 1 << 8
	ScreencapDerived                 = helperBase | 1
	ScreencapAlias   ScreencapMethod = ScreencapDerived
	ScreencapTop                     = 1 << 63
)

const (
	NegBase Signed = iota - 2
	NegNext
)

const ScreencapForward = ScreencapLater + 1

const ScreencapLater ScreencapMethod = 41
`,
			want: map[string]string{
				"ScreencapDerived": "257",
				"ScreencapAlias":   "257",
				"ScreencapTop":     "9223372036854775808",
				"NegBase":          "-2",
				"NegNext":          "-1",
				"ScreencapForward": "42",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evaluation, err := evaluateGoConstSource("fixture.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			if len(evaluation.failures) != 0 {
				t.Fatalf("unexpected failures: %+v", evaluation.failures)
			}
			for name, want := range tc.want {
				if got := constantGoExactString(t, evaluation, name); got != want {
					t.Errorf("constant %s: got %s want %s", name, got, want)
				}
			}
		})
	}
}

func TestEvalGoConstSource_DiagnosesUnsupportedConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		src          string
		wantFailures map[string]string
		wantValues   map[string]string
	}{
		{
			name:         "negative unsigned conversion",
			src:          "package fixture\n\nconst Broken = uint32(-1)\n",
			wantFailures: map[string]string{"Broken": "overflows"},
		},
		{
			name:         "typed unsigned overflow",
			src:          "package fixture\n\nconst Broken uint8 = 300\n",
			wantFailures: map[string]string{"Broken": "overflows"},
		},
		{
			name:         "negative unsigned constant",
			src:          "package fixture\n\nconst Broken uint8 = -1\n",
			wantFailures: map[string]string{"Broken": "overflows"},
		},
		{
			name:         "narrow typed value without conversion",
			src:          "package fixture\n\ntype ScreencapMethod uint64\n\nconst Broken ScreencapMethod = ^uint32(0)\n",
			wantFailures: map[string]string{"Broken": "cannot use"},
		},
		{
			name:         "typed shift overflow",
			src:          "package fixture\n\nconst Broken uint64 = 1 << 64\n",
			wantFailures: map[string]string{"Broken": "overflows"},
		},
		{
			name:         "unsupported function call",
			src:          "package fixture\n\nfunc helper() int { return 1 }\n\nconst Broken = helper()\n",
			wantFailures: map[string]string{"Broken": "undefined: helper"},
		},
		{
			name:         "missing dependency",
			src:          "package fixture\n\nconst Broken = Missing\n",
			wantFailures: map[string]string{"Broken": "undefined: Missing"},
		},
		{
			name:         "division by zero",
			src:          "package fixture\n\nconst Broken = 1 / 0\n",
			wantFailures: map[string]string{"Broken": "division by zero"},
		},
		{
			name: "dependency cycle",
			src:  "package fixture\n\nconst (\n\tBrokenA = BrokenB\n\tBrokenB = BrokenA\n)\n",
			wantFailures: map[string]string{
				"BrokenA": "initialization cycle",
				"BrokenB": "refers to BrokenA",
			},
		},
		{
			name:         "one failure does not hide independent constants",
			src:          "package fixture\n\nconst Good uint64 = uint64(^uint32(0))\n\nconst Broken = Missing\n",
			wantFailures: map[string]string{"Broken": "undefined: Missing"},
			wantValues:   map[string]string{"Good": "4294967295"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evaluation, err := evaluateGoConstSource("fixture.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			for name, wantErr := range tc.wantFailures {
				failure, ok := evaluation.failures[name]
				if !ok {
					t.Fatalf("expected failure for %s, got values=%+v failures=%+v", name, evaluation.values, evaluation.failures)
				}
				if !strings.Contains(failure.err, wantErr) {
					t.Fatalf("failure for %s: got %q want substring %q", name, failure.err, wantErr)
				}
			}
			for name, want := range tc.wantValues {
				if got := constantGoExactString(t, evaluation, name); got != want {
					t.Errorf("constant %s: got %s want %s", name, got, want)
				}
			}
		})
	}
}

func constantGoExactString(t *testing.T, evaluation *goConstEvaluation, name string) string {
	t.Helper()

	decl, ok := evaluation.values[name]
	if !ok {
		t.Fatalf("constant %s not evaluated; failures: %+v", name, evaluation.failures)
	}
	normalized, err := normalizeNumericValue(numericValue{kind: decl.kind, bits: decl.bits, val: decl.value})
	if err != nil {
		t.Fatalf("normalize %s: %v", name, err)
	}
	return normalized.ExactString()
}
