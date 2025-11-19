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
