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
