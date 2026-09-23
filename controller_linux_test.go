package maa

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLinuxController_InvalidConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux controller is only available on Linux")
	}

	ctrl, err := NewLinuxController(`{}`)
	require.Error(t, err)
	require.Nil(t, ctrl)
}
