package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolkit_ConfigInitOption(t *testing.T) {
	got := ConfigInitOption("./test", "{}")
	require.True(t, got)
}

func TestToolkit_FindAdbDevices(t *testing.T) {
	adbDevices := FindAdbDevices()
	require.NotNil(t, adbDevices)
}

func TestToolkit_FindDesktopWindows(t *testing.T) {
	desktopWindows := FindDesktopWindows()
	require.NotNil(t, desktopWindows)
}

type testToolKitRec struct{}

func (t *testToolKitRec) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return &CustomRecognitionResult{}, true
}

func TestToolkit_RegisterPICustomRecognition(t *testing.T) {
	PIRegisterCustomRecognition(0, "TestRec", &testToolKitRec{})
}

type testToolkitAct struct{}

func (t testToolkitAct) Run(_ *Context, _ *CustomActionArg) bool {
	return true
}

func TestToolkit_RegisterPICustomAction(t *testing.T) {
	PIRegisterCustomAction(0, "TestAct", &testToolkitAct{})
}

func TestToolkit_ClearPICustom(t *testing.T) {
	PIRegisterCustomRecognition(0, "TestRec", &testToolKitRec{})
	PIRegisterCustomAction(0, "TestAct", &testToolkitAct{})
	PIClearCustom(0)
}
