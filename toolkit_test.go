package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createToolkit(t *testing.T) *Toolkit {
	toolkit := NewToolkit()
	require.NotNil(t, toolkit)
	return toolkit
}

func TestNewToolkit(t *testing.T) {
	createToolkit(t)
}

func TestToolkit_ConfigInitOption(t *testing.T) {
	toolkit := createToolkit(t)
	got := toolkit.ConfigInitOption("./test", "{}")
	require.True(t, got)
}

func TestToolkit_FindAdbDevices(t *testing.T) {
	toolkit := createToolkit(t)
	adbDevices := toolkit.FindAdbDevices()
	require.NotNil(t, adbDevices)
}

func TestToolkit_FindDesktopWindows(t *testing.T) {
	toolkit := createToolkit(t)
	desktopWindows := toolkit.FindDesktopWindows()
	require.NotNil(t, desktopWindows)
}

type testToolKitRec struct{}

func (t *testToolKitRec) Run(_ *Context, _ *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return &CustomRecognitionResult{}, true
}

func TestToolkit_RegisterPICustomRecognition(t *testing.T) {
	toolkit := createToolkit(t)
	toolkit.RegisterPICustomRecognition(0, "TestRec", &testToolKitRec{})
}

type testToolkitAct struct{}

func (t testToolkitAct) Run(_ *Context, _ *CustomActionArg) bool {
	return true
}

func TestToolkit_RegisterPICustomAction(t *testing.T) {
	toolkit := createToolkit(t)
	toolkit.RegisterPICustomAction(0, "TestAct", &testToolkitAct{})
}

func TestToolkit_ClearPICustom(t *testing.T) {
	toolkit := createToolkit(t)
	toolkit.RegisterPICustomRecognition(0, "TestRec", &testToolKitRec{})
	toolkit.RegisterPICustomAction(0, "TestAct", &testToolkitAct{})
	toolkit.ClearPICustom(0)
}
