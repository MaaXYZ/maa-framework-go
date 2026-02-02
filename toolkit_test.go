package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolkit_ConfigInitOption(t *testing.T) {
	err := ConfigInitOption("./test", "{}")
	require.NoError(t, err)
}

func TestToolkit_FindAdbDevices(t *testing.T) {
	adbDevices, err := FindAdbDevices()
	require.NoError(t, err)
	require.NotNil(t, adbDevices)
}

func TestToolkit_FindDesktopWindows(t *testing.T) {
	desktopWindows, err := FindDesktopWindows()
	require.NoError(t, err)
	require.NotNil(t, desktopWindows)
}
