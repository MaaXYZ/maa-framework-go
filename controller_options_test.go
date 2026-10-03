package maa

import (
	"testing"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestController_SetBackgroundManagedKeys(t *testing.T) {
	ctrl := createBlankController(t)
	defer ctrl.Destroy()
	original := native.MaaControllerSetOption
	defer func() { native.MaaControllerSetOption = original }()

	for _, keys := range [][]int32{{0x41, 0x42, 0x100}, {}, nil} {
		calls := 0
		native.MaaControllerSetOption = func(handle uintptr, option native.MaaCtrlOption, value unsafe.Pointer, size uint64) bool {
			calls++
			require.Equal(t, ctrl.handle, handle)
			require.Equal(t, native.MaaCtrlOption_BackgroundManagedKeys, option)
			require.Equal(t, uint64(len(keys)*4), size)
			require.NotNil(t, value)
			if len(keys) > 0 {
				require.Equal(t, keys, unsafe.Slice((*int32)(value), len(keys)))
			}
			return true
		}
		require.NoError(t, ctrl.SetBackgroundManagedKeys(keys))
		require.Equal(t, 1, calls)
	}

	native.MaaControllerSetOption = func(uintptr, native.MaaCtrlOption, unsafe.Pointer, uint64) bool { return false }
	require.Error(t, ctrl.SetBackgroundManagedKeys([]int32{0x41}))
	require.NoError(t, ctrl.Destroy())
	native.MaaControllerSetOption = func(uintptr, native.MaaCtrlOption, unsafe.Pointer, uint64) bool {
		t.Fatal("closed controller called the native setter")
		return true
	}
	require.ErrorIs(t, ctrl.SetBackgroundManagedKeys(nil), ErrClosed)
}
