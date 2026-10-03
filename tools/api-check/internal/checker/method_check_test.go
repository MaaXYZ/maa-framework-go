package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCControllerMethodGroups_EvaluatesCompositeExpressions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	headerPath := filepath.Join(dir, "MaaDef.h")
	content := `#define MaaAdbScreencapMethod_None 0ULL
#define MaaAdbScreencapMethod_Encode 1ULL
#define MaaAdbScreencapMethod_All (~MaaAdbScreencapMethod_None)
#define MaaAdbScreencapMethod_Default \
	(MaaAdbScreencapMethod_All & (~MaaAdbScreencapMethod_Encode))
`
	if err := os.WriteFile(headerPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write header: %v", err)
	}

	groups, issues, err := parseCControllerMethodGroups(headerPath)
	if err != nil {
		t.Fatalf("parse c groups: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %+v", issues)
	}

	got := groups[methodGroupAdbScreencap]["Default"]
	want := ^uint64(1)
	if got != want {
		t.Fatalf("unexpected Default value: got=%d want=%d", got, want)
	}
}

func TestParseGoControllerMethodGroups_EvaluatesExpressionsAndReportsFailures(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goPath := filepath.Join(dir, "adb.go")
	src := `package adb

type ScreencapMethod uint64
type InputMethod uint64

const (
	ScreencapNone ScreencapMethod = 0
	ScreencapAll                 = ^ScreencapNone
	ScreencapBroken              = MissingConst

	InputNone InputMethod = 0
	InputAll             = ^InputNone
)
`
	if err := os.WriteFile(goPath, []byte(src), 0o600); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	groups, issues, err := parseGoControllerMethodGroups(goPath, "adb")
	if err != nil {
		t.Fatalf("parse go groups: %v", err)
	}
	if got, want := groups[methodGroupAdbScreencap]["All"], ^uint64(0); got != want {
		t.Fatalf("unexpected ScreencapAll: got=%d want=%d", got, want)
	}
	if got, want := groups[methodGroupAdbInput]["All"], ^uint64(0); got != want {
		t.Fatalf("unexpected InputAll: got=%d want=%d", got, want)
	}
	if !hasIssueMessageContaining(issues, "[adb.screencap] failed to evaluate Go method value: Broken") {
		t.Fatalf("expected Broken evaluation failure, got issues: %+v", issues)
	}
}

func TestParseGoControllerMethodGroups_NarrowConversionAndDependencies(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goPath := filepath.Join(dir, "adb.go")
	src := `package adb

type ScreencapMethod uint64
type InputMethod uint64

const (
	screencapHelper  ScreencapMethod = 1 << 8
	ScreencapNone    ScreencapMethod = 0
	ScreencapNarrow                  = uint64(^uint32(0))
	ScreencapWide    ScreencapMethod = ScreencapMethod(^uint32(0))
	ScreencapDerived                 = screencapHelper | 1
	ScreencapIotaBase ScreencapMethod = iota
	ScreencapIotaNext
	InputNone    InputMethod = 0
	InputShifted             = InputNone + (1 << 20)
)
`
	if err := os.WriteFile(goPath, []byte(src), 0o600); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	groups, issues, err := parseGoControllerMethodGroups(goPath, "adb")
	if err != nil {
		t.Fatalf("parse go groups: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %+v", issues)
	}

	want := map[string]map[string]uint64{
		methodGroupAdbScreencap: {
			"None":     0,
			"Narrow":   4294967295,
			"Wide":     4294967295,
			"Derived":  257,
			"IotaBase": 5,
			"IotaNext": 6,
		},
		methodGroupAdbInput: {
			"None":    0,
			"Shifted": 1 << 20,
		},
	}
	for group, methods := range want {
		for name, value := range methods {
			if got := groups[group][name]; got != value {
				t.Errorf("%s %s: got=%d want=%d", group, name, got, value)
			}
		}
	}
}

func TestParseGoControllerMethodGroups_DiagnosesOverflowAndUnsupportedCalls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goPath := filepath.Join(dir, "adb.go")
	src := `package adb

type ScreencapMethod uint64
type InputMethod uint64

func helper() uint64 { return 1 }

const (
	ScreencapOk       ScreencapMethod = 1
	ScreencapOverflow uint8 = 300
	ScreencapCall = helper()
	ScreencapCycleA = ScreencapCycleB
	ScreencapCycleB = ScreencapCycleA
)
`
	if err := os.WriteFile(goPath, []byte(src), 0o600); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	groups, issues, err := parseGoControllerMethodGroups(goPath, "adb")
	if err != nil {
		t.Fatalf("parse go groups: %v", err)
	}
	if got, want := groups[methodGroupAdbScreencap]["Ok"], uint64(1); got != want {
		t.Fatalf("unexpected ScreencapOk: got=%d want=%d", got, want)
	}
	for _, logicalName := range []string{"Overflow", "Call", "CycleA", "CycleB"} {
		if !hasIssueMessageContaining(issues, "[adb.screencap] failed to evaluate Go method value: "+logicalName) {
			t.Errorf("expected evaluation failure for %s, got issues: %+v", logicalName, issues)
		}
	}
}

func TestParseCControllerMethodGroups_HelperMacrosAndFailures(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	headerPath := filepath.Join(dir, "MaaDef.h")
	content := `#define MAA_ADB_HELPER (1ULL << 9)
#define MaaAdbScreencapMethod_None 0ULL
#define MaaAdbScreencapMethod_Derived (MAA_ADB_HELPER | 1ULL)
#define MaaAdbScreencapMethod_Undefined MissingHelper
#define MaaAdbScreencapMethod_CycleA MaaAdbScreencapMethod_CycleB
#define MaaAdbScreencapMethod_CycleB MaaAdbScreencapMethod_CycleA
#define MaaAdbScreencapMethod_Divide (1ULL / 0)

enum
{
    MAA_ADB_ENUM_HELPER = 7,
};
#define MaaAdbInputMethod_AdbShell MAA_ADB_ENUM_HELPER
`
	if err := os.WriteFile(headerPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write header: %v", err)
	}

	groups, issues, err := parseCControllerMethodGroups(headerPath)
	if err != nil {
		t.Fatalf("parse c groups: %v", err)
	}
	if got, want := groups[methodGroupAdbScreencap]["Derived"], uint64(513); got != want {
		t.Fatalf("unexpected Derived: got=%d want=%d", got, want)
	}
	if got, want := groups[methodGroupAdbInput]["AdbShell"], uint64(7); got != want {
		t.Fatalf("unexpected AdbShell: got=%d want=%d", got, want)
	}
	for _, logicalName := range []string{"Undefined", "CycleA", "CycleB", "Divide"} {
		if !hasIssueMessageContaining(issues, "[adb.screencap] failed to evaluate C method value: "+logicalName) {
			t.Errorf("expected evaluation failure for %s, got issues: %+v", logicalName, issues)
		}
	}
}

func TestCheckControllerMethodCoverage_DetectsMissingExtraMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	headerPath := filepath.Join(dir, "MaaDef.h")
	adbPath := filepath.Join(dir, "adb.go")
	win32Path := filepath.Join(dir, "win32.go")

	header := `#define MaaAdbScreencapMethod_None 0ULL
#define MaaAdbScreencapMethod_Encode 1ULL

#define MaaAdbInputMethod_None 0ULL
#define MaaAdbInputMethod_AdbShell 1ULL

#define MaaWin32ScreencapMethod_None 0ULL
#define MaaWin32ScreencapMethod_DXGI_DesktopDup (1ULL << 2)

#define MaaWin32InputMethod_None 0ULL
#define MaaWin32InputMethod_Seize 1ULL
`
	if err := os.WriteFile(headerPath, []byte(header), 0o600); err != nil {
		t.Fatalf("write header: %v", err)
	}

	adbSrc := `package adb

type ScreencapMethod uint64
type InputMethod uint64

const (
	ScreencapNone  ScreencapMethod = 0
	ScreencapEncode ScreencapMethod = 2
	ScreencapExtra ScreencapMethod = 8
	InputNone      InputMethod = 0
)
`
	if err := os.WriteFile(adbPath, []byte(adbSrc), 0o600); err != nil {
		t.Fatalf("write adb file: %v", err)
	}

	win32Src := `package win32

type ScreencapMethod uint64
type InputMethod uint64

const (
	ScreencapNone          ScreencapMethod = 0
	ScreencapDXGIDesktopDup ScreencapMethod = 1 << 2

	InputNone InputMethod = 0
	InputSeize InputMethod = 1
)
`
	if err := os.WriteFile(win32Path, []byte(win32Src), 0o600); err != nil {
		t.Fatalf("write win32 file: %v", err)
	}

	issues, err := checkControllerMethodCoverage(headerPath, adbPath, win32Path)
	if err != nil {
		t.Fatalf("check method coverage: %v", err)
	}

	assertIssueContains(t, issues, "[adb.screencap] method value mismatch: Encode")
	assertIssueContains(t, issues, "[adb.screencap] Go method not found in C: Extra")
	assertIssueContains(t, issues, "[adb.input] C method not found in Go: AdbShell")
	assertIssueNotContains(t, issues, "[win32.screencap] C method not found in Go: DXGI_DesktopDup")
	assertIssueNotContains(t, issues, "[win32.screencap] Go method not found in C: DXGIDesktopDup")
}

func TestCompareMethodGroupValues_CSideUnderscoreNormalizationCollision(t *testing.T) {
	t.Parallel()

	cValues := map[string]uint64{
		"A_B": 1,
		"AB":  2,
	}
	goValues := map[string]uint64{
		"AB": 1,
	}

	issues := compareMethodGroupValues(methodGroupWin32Screencap, cValues, goValues)

	assertIssueContains(t, issues, "[win32.screencap] C method name collision after underscore normalization: AB => AB(c=2), A_B(c=1)")
	assertIssueNotContains(t, issues, "[win32.screencap] Go method not found in C: AB")
	assertIssueNotContains(t, issues, "[win32.screencap] C method not found in Go: ")
}

func hasIssueMessageContaining(issues []issue, needle string) bool {
	for _, it := range issues {
		if strings.Contains(it.message, needle) {
			return true
		}
	}
	return false
}

func assertIssueContains(t *testing.T, issues []issue, needle string) {
	t.Helper()
	if !hasIssueMessageContaining(issues, needle) {
		t.Fatalf("expected issue containing %q, got: %+v", needle, issues)
	}
}

func assertIssueNotContains(t *testing.T, issues []issue, needle string) {
	t.Helper()
	if hasIssueMessageContaining(issues, needle) {
		t.Fatalf("expected no issue containing %q, got: %+v", needle, issues)
	}
}
