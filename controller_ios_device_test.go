package maa

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/ios"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/store"
	"github.com/stretchr/testify/require"
)

func TestNewIOSController_ValidationAndNativeFailure(t *testing.T) {
	calls := 0
	replaceNativeForTest(t, &native.MaaIOSControllerCreate, func(udid string, methods uint64, connection int32, config string) uintptr {
		calls++
		require.Equal(t, "device-udid", udid)
		require.Equal(t, ^uint64(0), methods)
		require.Equal(t, int32(ios.ConnectionWiFi), connection)
		require.Equal(t, `{"wifi_address":"192.0.2.1"}`, config)
		return 0
	})
	for _, udid := range []string{"", "device\x00suffix"} {
		ctrl, err := NewIOSController(udid, ios.ScreencapDefault, ios.ConnectionAuto, "")
		require.Error(t, err)
		require.Nil(t, ctrl)
	}
	ctrl, err := NewIOSController("device-udid", ios.ScreencapDefault, ios.ConnectionAuto, "{}\x00ignored")
	require.Error(t, err)
	require.Nil(t, ctrl)
	require.Zero(t, calls)

	before := liveNativeObjects.Load()
	ctrl, err = NewIOSController("device-udid", ios.ScreencapDefault, ios.ConnectionWiFi, `{"wifi_address":"192.0.2.1"}`)
	require.Error(t, err)
	require.Nil(t, ctrl)
	require.Equal(t, 1, calls)
	require.Equal(t, before, liveNativeObjects.Load())
}

func TestNewIOSController_OwnedLifecycle(t *testing.T) {
	const handle = uintptr(0x7654321)
	destroyed := 0
	replaceNativeForTest(t, &native.MaaIOSControllerCreate, func(_ string, _ uint64, _ int32, config string) uintptr {
		require.Empty(t, config, "empty configuration must keep native defaults")
		return handle
	})
	replaceNativeForTest(t, &native.MaaControllerDestroy, func(got uintptr) {
		require.Equal(t, handle, got)
		destroyed++
	})
	before := liveNativeObjects.Load()
	ctrl, err := NewIOSController("device", ios.ScreencapScreenshotService, ios.ConnectionUSB, "")
	require.NoError(t, err)
	require.True(t, ctrl.owned)
	require.Equal(t, before+1, liveNativeObjects.Load())
	store.CtrlStore.Update(handle, func(value *store.CtrlStoreValue) { require.NotNil(t, value.SinkIDToEventCallbackID) })
	state, ok := controllerStates.Load(handle)
	require.True(t, ok)
	require.Same(t, ctrl.state, state)
	require.NoError(t, ctrl.Destroy())
	require.NoError(t, ctrl.Destroy())
	require.Equal(t, 1, destroyed)
	require.Equal(t, before, liveNativeObjects.Load())
	_, ok = controllerStates.Load(handle)
	require.False(t, ok)
}

func TestNewIOSController_RealNativeLifecycle(t *testing.T) {
	before := liveNativeObjects.Load()
	ctrl, err := NewIOSController("go-binding-test-device", ios.ScreencapScreenshotService, ios.ConnectionAuto, "{}")
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	t.Cleanup(func() { require.NoError(t, ctrl.Destroy()) })
	require.False(t, ctrl.Connected(), "construction must not connect to a device")
	require.Equal(t, before+1, liveNativeObjects.Load())
	require.NoError(t, ctrl.Destroy())
	require.NoError(t, ctrl.Destroy())
	require.Equal(t, before, liveNativeObjects.Load())
}
