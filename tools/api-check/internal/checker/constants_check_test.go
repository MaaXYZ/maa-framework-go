package checker

import (
	"fmt"
	"go/constant"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstantNormalization_RejectsGoUnsignedNegatives(t *testing.T) {
	t.Parallel()
	value := numericValue{kind: numericUnsigned, bits: 32, val: constant.MakeInt64(-1)}
	if _, err := normalizeNumericValue(value); err == nil {
		t.Fatal("expected unsigned negative to be rejected outside C conversion")
	}
	cValue, err := normalizeCNumericValue(value)
	if err != nil {
		t.Fatalf("C conversion: %v", err)
	}
	if got := cValue.ExactString(); got != "4294967295" {
		t.Fatalf("C conversion: got %s want 4294967295", got)
	}
}

func TestCheckConstantCoverage_EmptyFamilyFails(t *testing.T) {
	t.Parallel()
	spec := constantFamilySpecs[0]
	issues := compareConstantFixture(t, spec, "", "")
	assertConstantIssueContaining(t, issues, "[%s] no C constants found", spec.name)
	assertConstantIssueContaining(t, issues, "[%s] no Go constants found", spec.name)
}

// constantFamilyFixture is the self-contained cross-language fixture for one
// inventory family. It mirrors the declaration shape of the real sources so
// the family comparison, normalization, and alias rules are exercised without
// native libraries.
type constantFamilyFixture struct {
	cHeader  string
	cSource  string
	goSource string

	// Mutation anchors. goRemove/cRemove are substrings whose whole line is
	// deleted; cMismatchFrom/To change one C value.
	goRemove      string
	cRemove       string
	cMismatchFrom string
	cMismatchTo   string

	cMissingInGoLogical string
	goMissingInCLogical string
	mismatchName        string
}

var constantFamilyFixtures = map[string]constantFamilyFixture{
	"gamepad.type": {
		cHeader:  maaDefHeaderRel,
		cSource:  "#define MaaGamepadType_Xbox360 0ULL\n#define MaaGamepadType_DualShock4 1ULL\n",
		goSource: "type GamepadType uint64\nconst GamepadTypeXbox360 GamepadType = 0\nconst GamepadTypeDualShock4 GamepadType = 1\n",
		goRemove: "GamepadTypeDualShock4", cRemove: "#define MaaGamepadType_DualShock4",
		cMismatchFrom: "#define MaaGamepadType_DualShock4 1ULL", cMismatchTo: "#define MaaGamepadType_DualShock4 2ULL",
		cMissingInGoLogical: "DualShock4", goMissingInCLogical: "DualShock4", mismatchName: "DualShock4",
	},
	"macos.permission": {
		cHeader:  maaToolkitDefHeaderRel,
		cSource:  "#define MaaMacOSPermissionScreenCapture 1\n#define MaaMacOSPermissionAccessibility 2\n",
		goSource: "type MacOSPermission int32\nconst MacOSPermissionScreenCapture MacOSPermission = 1\nconst MacOSPermissionAccessibility MacOSPermission = 2\n",
		goRemove: "MacOSPermissionAccessibility", cRemove: "#define MaaMacOSPermissionAccessibility",
		cMismatchFrom: "#define MaaMacOSPermissionAccessibility 2", cMismatchTo: "#define MaaMacOSPermissionAccessibility 3",
		cMissingInGoLogical: "Accessibility", goMissingInCLogical: "Accessibility", mismatchName: "Accessibility",
	},
	"macos.screencap": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaMacOSScreencapMethod_None 0ULL\n" +
			"#define MaaMacOSScreencapMethod_ScreenCaptureKit 1ULL\n",
		goSource: `type ScreencapMethod uint64

const (
	ScreencapNone             ScreencapMethod = 0
	ScreencapScreenCaptureKit ScreencapMethod = 1
)
`,
		goRemove:            "ScreencapScreenCaptureKit",
		cRemove:             "#define MaaMacOSScreencapMethod_ScreenCaptureKit",
		cMismatchFrom:       "#define MaaMacOSScreencapMethod_ScreenCaptureKit 1ULL",
		cMismatchTo:         "#define MaaMacOSScreencapMethod_ScreenCaptureKit 2ULL",
		cMissingInGoLogical: "ScreenCaptureKit",
		goMissingInCLogical: "ScreenCaptureKit",
		mismatchName:        "ScreenCaptureKit",
	},
	"macos.input": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaMacOSInputMethod_None 0ULL\n" +
			"#define MaaMacOSInputMethod_GlobalEvent 1ULL\n" +
			"#define MaaMacOSInputMethod_PostToPid (1ULL << 1)\n",
		goSource: `type InputMethod uint64

const (
	InputNone        InputMethod = 0
	InputGlobalEvent InputMethod = 1
	InputPostToPid   InputMethod = 1 << 1
)
`,
		goRemove:            "InputPostToPid",
		cRemove:             "#define MaaMacOSInputMethod_PostToPid",
		cMismatchFrom:       "#define MaaMacOSInputMethod_PostToPid (1ULL << 1)",
		cMismatchTo:         "#define MaaMacOSInputMethod_PostToPid (1ULL << 2)",
		cMissingInGoLogical: "PostToPid",
		goMissingInCLogical: "PostToPid",
		mismatchName:        "PostToPid",
	},
	"native.gamepad_type": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaGamepadType_Xbox360 0ULL\n" +
			"#define MaaGamepadType_DualShock4 1ULL\n",
		goSource: `type MaaGamepadType uint64

const (
	MaaGamepadType_Xbox360    MaaGamepadType = 0
	MaaGamepadType_DualShock4 MaaGamepadType = 1
)
`,
		goRemove:            "MaaGamepadType_DualShock4",
		cRemove:             "#define MaaGamepadType_DualShock4",
		cMismatchFrom:       "#define MaaGamepadType_DualShock4 1ULL",
		cMismatchTo:         "#define MaaGamepadType_DualShock4 2ULL",
		cMissingInGoLogical: "DualShock4",
		goMissingInCLogical: "DualShock4",
		mismatchName:        "DualShock4",
	},
	"gamepad.button": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaGamepadButton_A 0x1000\n" +
			"#define MaaGamepadButton_LEFT_THUMB 0x0040\n" +
			"#define MaaGamepadButton_CROSS MaaGamepadButton_A\n" +
			"#define MaaGamepadButton_PS 0x10000\n",
		goSource: `type Button int32

const (
	ButtonA         Button = 0x1000
	ButtonLeftThumb Button = 0x0040
	ButtonCross     Button = ButtonA
	ButtonPS        Button = 0x10000
)
`,
		goRemove:            "ButtonLeftThumb",
		cRemove:             "#define MaaGamepadButton_LEFT_THUMB",
		cMismatchFrom:       "#define MaaGamepadButton_PS 0x10000",
		cMismatchTo:         "#define MaaGamepadButton_PS 0x20000",
		cMissingInGoLogical: "LEFT_THUMB",
		goMissingInCLogical: "LeftThumb",
		mismatchName:        "PS",
	},
	"gamepad.touch": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaGamepadTouch_LeftStick 0\n" +
			"#define MaaGamepadTouch_LeftTrigger 2\n" +
			"#define MaaGamepadTouch_RightTrigger 3\n",
		goSource: `type Touch int32

const (
	TouchLeftStick    Touch = 0
	TouchLeftTrigger  Touch = 2
	TouchRightTrigger Touch = 3
)
`,
		goRemove:            "TouchRightTrigger",
		cRemove:             "#define MaaGamepadTouch_RightTrigger",
		cMismatchFrom:       "#define MaaGamepadTouch_RightTrigger 3",
		cMismatchTo:         "#define MaaGamepadTouch_RightTrigger 4",
		cMissingInGoLogical: "RightTrigger",
		goMissingInCLogical: "RightTrigger",
		mismatchName:        "RightTrigger",
	},
	"controller.feature": {
		cHeader: maaDefHeaderRel,
		cSource: "#define MaaControllerFeature_None 0\n" +
			"#define MaaControllerFeature_UseMouseDownAndUpInsteadOfClick 1ULL\n" +
			"#define MaaControllerFeature_UseKeyboardDownAndUpInsteadOfClick (1ULL << 1)\n" +
			"#define MaaControllerFeature_NoScalingTouchPoints (1ULL << 2)\n",
		goSource: `type ControllerFeature uint64

const (
	ControllerFeatureNone                               ControllerFeature = 0
	ControllerFeatureUseMouseDownAndUpInsteadOfClick    ControllerFeature = 1
	ControllerFeatureUseKeyboardDownAndUpInsteadOfClick ControllerFeature = 1 << 1
	ControllerFeatureNoScalingTouchPoints               ControllerFeature = 1 << 2
)
`,
		goRemove:            "ControllerFeatureNoScalingTouchPoints",
		cRemove:             "#define MaaControllerFeature_NoScalingTouchPoints",
		cMismatchFrom:       "#define MaaControllerFeature_NoScalingTouchPoints (1ULL << 2)",
		cMismatchTo:         "#define MaaControllerFeature_NoScalingTouchPoints (1ULL << 3)",
		cMissingInGoLogical: "NoScalingTouchPoints",
		goMissingInCLogical: "NoScalingTouchPoints",
		mismatchName:        "NoScalingTouchPoints",
	},
	"status": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaStatusEnum
{
    MaaStatus_Invalid = 0,
    MaaStatus_Pending = 1000,
    MaaStatus_Running = 2000,
    MaaStatus_Succeeded = 3000,
    MaaStatus_Failed = 4000,
};
`,
		goSource: `type Status int32

const (
	StatusInvalid Status = 0
	StatusPending Status = 1000
	StatusRunning Status = 2000
	StatusSuccess Status = 3000
	StatusFailure Status = 4000
)
`,
		goRemove:            "StatusPending",
		cRemove:             "MaaStatus_Pending = 1000",
		cMismatchFrom:       "MaaStatus_Pending = 1000",
		cMismatchTo:         "MaaStatus_Pending = 1001",
		cMissingInGoLogical: "Pending",
		goMissingInCLogical: "Pending",
		mismatchName:        "Pending",
	},
	"logging_level": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaLoggingLevelEnum
{
    MaaLoggingLevel_Off = 0,
    MaaLoggingLevel_Fatal = 1,
    MaaLoggingLevel_Error = 2,
    MaaLoggingLevel_Warn = 3,
    MaaLoggingLevel_Info = 4,
    MaaLoggingLevel_Debug = 5,
    MaaLoggingLevel_Trace = 6,
    MaaLoggingLevel_All = 7,
};
`,
		goSource: `type LoggingLevel int32

const (
	LoggingLevelOff LoggingLevel = iota
	LoggingLevelFatal
	LoggingLevelError
	LoggingLevelWarn
	LoggingLevelInfo
	LoggingLevelDebug
	LoggingLevelTrace
	LoggingLevelAll
)
`,
		goRemove:            "LoggingLevelAll",
		cRemove:             "MaaLoggingLevel_All = 7",
		cMismatchFrom:       "MaaLoggingLevel_All = 7",
		cMismatchTo:         "MaaLoggingLevel_All = 8",
		cMissingInGoLogical: "All",
		goMissingInCLogical: "All",
		mismatchName:        "All",
	},
	"native.global_option": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaGlobalOptionEnum
{
    MaaGlobalOption_Invalid = 0,
    MaaGlobalOption_LogDir = 1,
    MaaGlobalOption_StdoutLevel = 4,
    MaaGlobalOption_RecoImageCacheLimit = 9,
};
`,
		goSource: `type MaaGlobalOption int32

const (
	MaaGlobalOption_Invalid              MaaGlobalOption = 0
	MaaGlobalOption_LogDir               MaaGlobalOption = 1
	MaaGlobalOption_StdoutLevel          MaaGlobalOption = 4
	MaaGlobalOption_RecoImageCacheLimit  MaaGlobalOption = 9
)
`,
		goRemove:            "MaaGlobalOption_RecoImageCacheLimit",
		cRemove:             "MaaGlobalOption_RecoImageCacheLimit = 9",
		cMismatchFrom:       "MaaGlobalOption_RecoImageCacheLimit = 9",
		cMismatchTo:         "MaaGlobalOption_RecoImageCacheLimit = 10",
		cMissingInGoLogical: "RecoImageCacheLimit",
		goMissingInCLogical: "RecoImageCacheLimit",
		mismatchName:        "RecoImageCacheLimit",
	},
	"native.res_option": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaResOptionEnum
{
    MaaResOption_Invalid = 0,
    MaaResOption_InferenceDevice = 1,
    MaaResOption_InferenceExecutionProvider = 2,
};
`,
		goSource: `type MaaResOption int32

const (
	MaaResOption_Invalid                    MaaResOption = 0
	MaaResOption_InferenceDevice            MaaResOption = 1
	MaaResOption_InferenceExecutionProvider MaaResOption = 2
)
`,
		goRemove:            "MaaResOption_InferenceExecutionProvider",
		cRemove:             "MaaResOption_InferenceExecutionProvider = 2",
		cMismatchFrom:       "MaaResOption_InferenceExecutionProvider = 2",
		cMismatchTo:         "MaaResOption_InferenceExecutionProvider = 3",
		cMissingInGoLogical: "InferenceExecutionProvider",
		goMissingInCLogical: "InferenceExecutionProvider",
		mismatchName:        "InferenceExecutionProvider",
	},
	"native.ctrl_option": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaCtrlOptionEnum
{
    MaaCtrlOption_Invalid = 0,
    MaaCtrlOption_MouseLockFollow = 4,
    MaaCtrlOption_ScreenshotResizeMethod = 6,
    MaaCtrlOption_ScreenshotTargetExpand = 8,
};
`,
		goSource: `type MaaCtrlOption int32

const (
	MaaCtrlOption_Invalid                MaaCtrlOption = 0
	MaaCtrlOption_MouseLockFollow        MaaCtrlOption = 4
	MaaCtrlOption_ScreenshotResizeMethod MaaCtrlOption = 6
	MaaCtrlOption_ScreenshotTargetExpand MaaCtrlOption = 8
)
`,
		goRemove:            "MaaCtrlOption_ScreenshotTargetExpand",
		cRemove:             "MaaCtrlOption_ScreenshotTargetExpand = 8",
		cMismatchFrom:       "MaaCtrlOption_ScreenshotTargetExpand = 8",
		cMismatchTo:         "MaaCtrlOption_ScreenshotTargetExpand = 9",
		cMissingInGoLogical: "ScreenshotTargetExpand",
		goMissingInCLogical: "ScreenshotTargetExpand",
		mismatchName:        "ScreenshotTargetExpand",
	},
	"native.tasker_option": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaTaskerOptionEnum
{
    MaaTaskerOption_Invalid = 0,
};
`,
		goSource: `type MaaTaskerOption int32

const (
	MaaTaskerOption_Invalid MaaTaskerOption = 0
)
`,
		goRemove:            "MaaTaskerOption_Invalid",
		cRemove:             "MaaTaskerOption_Invalid = 0",
		cMismatchFrom:       "MaaTaskerOption_Invalid = 0",
		cMismatchTo:         "MaaTaskerOption_Invalid = 1",
		cMissingInGoLogical: "Invalid",
		goMissingInCLogical: "Invalid",
		mismatchName:        "Invalid",
	},
	"native.inference_device": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaInferenceDeviceEnum
{
    MaaInferenceDevice_CPU = -2,
    MaaInferenceDevice_Auto = -1,
    MaaInferenceDevice_0 = 0,
    MaaInferenceDevice_1 = 1,
};
`,
		goSource: `type MaaInferenceDevice int32

const (
	MaaInferenceDevice_CPU  MaaInferenceDevice = -2
	MaaInferenceDevice_Auto MaaInferenceDevice = -1
	MaaInferenceDevice_0    MaaInferenceDevice = 0
	MaaInferenceDevice_1    MaaInferenceDevice = 1
)
`,
		goRemove:            "MaaInferenceDevice_1",
		cRemove:             "MaaInferenceDevice_1 = 1",
		cMismatchFrom:       "MaaInferenceDevice_CPU = -2",
		cMismatchTo:         "MaaInferenceDevice_CPU = -3",
		cMissingInGoLogical: "1",
		goMissingInCLogical: "1",
		mismatchName:        "CPU",
	},
	"native.execution_provider": {
		cHeader: maaDefHeaderRel,
		cSource: `enum MaaInferenceExecutionProviderEnum
{
    MaaInferenceExecutionProvider_Auto = 0,
    MaaInferenceExecutionProvider_CPU = 1,
    MaaInferenceExecutionProvider_CUDA = 4,
};
`,
		goSource: `type MaaInferenceExecutionProvider int32

const (
	MaaInferenceExecutionProvider_Auto MaaInferenceExecutionProvider = 0
	MaaInferenceExecutionProvider_CPU  MaaInferenceExecutionProvider = 1
	MaaInferenceExecutionProvider_CUDA MaaInferenceExecutionProvider = 4
)
`,
		goRemove:            "MaaInferenceExecutionProvider_CUDA",
		cRemove:             "MaaInferenceExecutionProvider_CUDA = 4",
		cMismatchFrom:       "MaaInferenceExecutionProvider_CUDA = 4",
		cMismatchTo:         "MaaInferenceExecutionProvider_CUDA = 5",
		cMissingInGoLogical: "CUDA",
		goMissingInCLogical: "CUDA",
		mismatchName:        "CUDA",
	},
	"native.macos_permission": {
		cHeader: maaToolkitDefHeaderRel,
		cSource: `enum MaaMacOSPermissionEnum
{
    MaaMacOSPermissionScreenCapture = 1,
    MaaMacOSPermissionAccessibility = 2,
};
`,
		goSource: `type MaaMacOSPermission int32

const (
	MaaMacOSPermissionScreenCapture MaaMacOSPermission = 1
	MaaMacOSPermissionAccessibility MaaMacOSPermission = 2
)
`,
		goRemove:            "MaaMacOSPermissionAccessibility",
		cRemove:             "MaaMacOSPermissionAccessibility = 2",
		cMismatchFrom:       "MaaMacOSPermissionScreenCapture = 1",
		cMismatchTo:         "MaaMacOSPermissionScreenCapture = 3",
		cMissingInGoLogical: "Accessibility",
		goMissingInCLogical: "Accessibility",
		mismatchName:        "ScreenCapture",
	},
}

func TestCheckConstantCoverage_FamilyMutations(t *testing.T) {
	t.Parallel()

	if len(constantFamilyFixtures) != len(constantFamilySpecs) {
		t.Fatalf("fixture count %d does not match family count %d", len(constantFamilyFixtures), len(constantFamilySpecs))
	}

	for _, spec := range constantFamilySpecs {
		fixture, ok := constantFamilyFixtures[spec.name]
		if !ok {
			t.Fatalf("missing fixture for family %s", spec.name)
		}

		t.Run(spec.name, func(t *testing.T) {
			t.Run("baseline", func(t *testing.T) {
				issues := compareConstantFixture(t, spec, fixture.cSource, fixture.goSource)
				if len(issues) != 0 {
					t.Fatalf("expected clean family, got %+v", issues)
				}
			})

			t.Run("c constant missing in go", func(t *testing.T) {
				goSource := removeConstantLine(t, fixture.goSource, fixture.goRemove)
				issues := compareConstantFixture(t, spec, fixture.cSource, goSource)
				assertConstantIssueContaining(t, issues, "[%s] C constant not found in Go: %s", spec.name, fixture.cMissingInGoLogical)
			})

			t.Run("go constant missing in c", func(t *testing.T) {
				cSource := removeConstantLine(t, fixture.cSource, fixture.cRemove)
				issues := compareConstantFixture(t, spec, cSource, fixture.goSource)
				assertConstantIssueContaining(t, issues, "[%s] Go constant not found in C: %s", spec.name, fixture.goMissingInCLogical)
			})

			t.Run("value mismatch", func(t *testing.T) {
				cSource := replaceConstantOnce(t, fixture.cSource, fixture.cMismatchFrom, fixture.cMismatchTo)
				issues := compareConstantFixture(t, spec, cSource, fixture.goSource)
				assertConstantIssueContaining(t, issues, "[%s] constant value mismatch: %s", spec.name, fixture.mismatchName)
			})
		})
	}
}

func TestCheckConstantCoverage_DeliberateAliases(t *testing.T) {
	t.Parallel()

	statusSpec := constantFamilySpec{
		name:     "status",
		goFile:   "status.go",
		goPrefix: "Status",
		cPrefix:  "MaaStatus_",
		kind:     numericSigned,
		bits:     32,
		aliases:  map[string]string{"Succeeded": "Success", "Failed": "Failure"},
	}
	fixture := constantFamilyFixtures["status"]

	t.Run("aliased members match by name", func(t *testing.T) {
		issues := compareConstantFixture(t, statusSpec, fixture.cSource, fixture.goSource)
		if len(issues) != 0 {
			t.Fatalf("expected aliases to match, got %+v", issues)
		}
	})

	t.Run("aliased mismatch names both sides", func(t *testing.T) {
		goSource := replaceConstantOnce(t, fixture.goSource, "StatusSuccess Status = 3000", "StatusSuccess Status = 3001")
		issues := compareConstantFixture(t, statusSpec, fixture.cSource, goSource)
		assertConstantIssueContaining(t, issues, "[status] constant value mismatch: Succeeded -> Success")
	})

	t.Run("removed aliased go constant is still reported", func(t *testing.T) {
		goSource := removeConstantLine(t, fixture.goSource, "StatusSuccess")
		issues := compareConstantFixture(t, statusSpec, fixture.cSource, goSource)
		assertConstantIssueContaining(t, issues, "[status] C constant not found in Go: Succeeded")
	})

	t.Run("gamepad ds4 alias is not collapsed by equal value", func(t *testing.T) {
		gamepadSpec := constantFamilySpec{
			name:     "gamepad.button",
			goFile:   "controller/gamepad/gamepad.go",
			goPrefix: "Button",
			cPrefix:  "MaaGamepadButton_",
			kind:     numericUnsigned,
			bits:     64,
			foldCase: true,
		}
		gamepadFixture := constantFamilyFixtures["gamepad.button"]
		if issues := compareConstantFixture(t, gamepadSpec, gamepadFixture.cSource, gamepadFixture.goSource); len(issues) != 0 {
			t.Fatalf("expected gamepad fixture to match, got %+v", issues)
		}
		goSource := removeConstantLine(t, gamepadFixture.goSource, "ButtonCross")
		issues := compareConstantFixture(t, gamepadSpec, gamepadFixture.cSource, goSource)
		assertConstantIssueContaining(t, issues, "[gamepad.button] C constant not found in Go: CROSS")
	})
}

func TestCheckConstantCoverage_NameCollision(t *testing.T) {
	t.Parallel()

	spec := constantFamilySpec{
		name:     "macos.screencap",
		goFile:   "controller/macos/macos.go",
		goPrefix: "Screencap",
		cPrefix:  "MaaMacOSScreencapMethod_",
		kind:     numericUnsigned,
		bits:     64,
	}
	cSource := "#define MaaMacOSScreencapMethod_A_B 1ULL\n#define MaaMacOSScreencapMethod_AB 2ULL\n"
	goSource := "type ScreencapMethod uint64\n\nconst ScreencapAB ScreencapMethod = 1\n"

	issues := compareConstantFixture(t, spec, cSource, goSource)
	assertConstantIssueContaining(t, issues, "[macos.screencap] C constant name collision after normalization: AB => AB(C=2), A_B(C=1)")
	assertConstantIssueNotContaining(t, issues, "[macos.screencap] Go constant not found in C: AB")
}

func TestCheckConstantCoverage_FixtureTree(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	headerDir := t.TempDir()
	writeConstantCoverageFixtureTree(t, repoRoot, headerDir)

	issues, err := checkConstantCoverage(repoRoot, headerDir)
	if err != nil {
		t.Fatalf("checkConstantCoverage: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected clean fixture tree, got %+v", issues)
	}

	macosPath := filepath.Join(repoRoot, "controller", "macos", "macos.go")
	content, err := os.ReadFile(macosPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	mutated := removeConstantLine(t, string(content), "ScreencapScreenCaptureKit")
	constantWriteFixtureFile(t, macosPath, mutated)

	issues, err = checkConstantCoverage(repoRoot, headerDir)
	if err != nil {
		t.Fatalf("checkConstantCoverage after mutation: %v", err)
	}
	assertConstantIssueContaining(t, issues, "[macos.screencap] C constant not found in Go: ScreenCaptureKit")
}

func TestCheckConstantCoverage_InputErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing header", func(t *testing.T) {
		if _, err := checkConstantCoverage(t.TempDir(), t.TempDir()); err == nil {
			t.Fatal("expected an error for missing headers")
		}
	})

	t.Run("missing go source", func(t *testing.T) {
		repoRoot := t.TempDir()
		headerDir := t.TempDir()
		constantWriteFixtureFile(t, filepath.Join(headerDir, maaDefHeaderRel), "#define MaaMacOSScreencapMethod_None 0ULL\n")
		constantWriteFixtureFile(t, filepath.Join(headerDir, maaToolkitDefHeaderRel), "")
		if _, err := checkConstantCoverage(repoRoot, headerDir); err == nil {
			t.Fatal("expected an error for missing Go sources")
		}
	})
}

func compareConstantFixture(t *testing.T, spec constantFamilySpec, cContent string, goContent string) []issue {
	t.Helper()

	env, err := evaluateCConstSources([]cConstSource{{path: "fixture.h", content: cContent}})
	if err != nil {
		t.Fatalf("evaluate C fixture: %v", err)
	}
	evaluation, err := evaluateGoConstSource("fixture.go", []byte("package fixture\n\n"+goContent))
	if err != nil {
		t.Fatalf("evaluate Go fixture: %v", err)
	}
	return compareConstantFamily(spec, env, evaluation)
}

func removeConstantLine(t *testing.T, content string, needle string) string {
	t.Helper()

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	removed := false
	for _, line := range lines {
		if !removed && strings.Contains(line, needle) {
			removed = true
			continue
		}
		out = append(out, line)
	}
	if !removed {
		t.Fatalf("needle %q not found in fixture:\n%s", needle, content)
	}
	return strings.Join(out, "\n")
}

func replaceConstantOnce(t *testing.T, content string, from string, to string) string {
	t.Helper()

	if !strings.Contains(content, from) {
		t.Fatalf("replacement source %q not found in fixture:\n%s", from, content)
	}
	return strings.Replace(content, from, to, 1)
}

func writeConstantCoverageFixtureTree(t *testing.T, repoRoot string, headerDir string) {
	t.Helper()

	goSnippets := map[string][]string{}
	cSnippets := map[string][]string{}
	for _, spec := range constantFamilySpecs {
		fixture, ok := constantFamilyFixtures[spec.name]
		if !ok {
			t.Fatalf("missing fixture for family %s", spec.name)
		}
		goSnippets[spec.goFile] = append(goSnippets[spec.goFile], fixture.goSource)
		cSnippets[fixture.cHeader] = append(cSnippets[fixture.cHeader], fixture.cSource)
	}

	for rel, snippets := range goSnippets {
		content := "package " + constantFixturePackage(rel) + "\n\n" + strings.Join(snippets, "\n")
		constantWriteFixtureFile(t, filepath.Join(repoRoot, rel), content)
	}
	for rel, snippets := range cSnippets {
		constantWriteFixtureFile(t, filepath.Join(headerDir, rel), strings.Join(snippets, "\n"))
	}
}

func constantFixturePackage(rel string) string {
	dir := filepath.Dir(rel)
	if dir == "." {
		return "maa"
	}
	return filepath.Base(dir)
}

func constantWriteFixtureFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertConstantIssueContaining(t *testing.T, issues []issue, format string, args ...any) {
	t.Helper()

	needle := fmt.Sprintf(format, args...)
	if !hasIssueMessageContaining(issues, needle) {
		t.Fatalf("expected issue containing %q, got: %+v", needle, issues)
	}
}

func assertConstantIssueNotContaining(t *testing.T, issues []issue, format string, args ...any) {
	t.Helper()

	needle := fmt.Sprintf(format, args...)
	if hasIssueMessageContaining(issues, needle) {
		t.Fatalf("expected no issue containing %q, got: %+v", needle, issues)
	}
}
