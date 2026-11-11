package maa

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/ios"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func iosDiscoveryCancellationCount() int {
	iosDiscoveryCancellations.RLock()
	defer iosDiscoveryCancellations.RUnlock()
	return len(iosDiscoveryCancellations.entries)
}

func stubIOSDiscovery(t *testing.T) *atomic.Int64 {
	t.Helper()
	var destroyed atomic.Int64
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListCreate, func() uintptr { return 42 })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListDestroy, func(uintptr) { destroyed.Add(1) })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFind, func(uintptr) bool { return true })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(int32, string, native.MaaToolkitIOSDeviceCancelCallbackPtr, uintptr, uintptr) bool { return true })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListSize, func(uintptr) uint64 { return 0 })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListGetWarnings, func(uintptr) string { return `["partial discovery"]` })
	return &destroyed
}

func TestToolkit_IOSDiscoverySnapshot(t *testing.T) {
	destroyed := stubIOSDiscovery(t)
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(int32, string, native.MaaToolkitIOSDeviceCancelCallbackPtr, uintptr, uintptr) bool {
		t.Error("default discovery must use MaaToolkitIOSDeviceFind")
		return false
	})
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListSize, func(uintptr) uint64 { return 1 })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListAt, func(list uintptr, index uint64) uintptr {
		require.Equal(t, uintptr(42), list)
		require.Zero(t, index)
		return 43
	})
	connections := `[{"transport":"usb","paired":true},{"transport":"wifi","address":"192.0.2.1","port":49152,"interface":"en0","extra":"preserved"}]`
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceGetName, func(device uintptr) string {
		require.Zero(t, destroyed.Load(), "fields must be copied before list destruction")
		require.Equal(t, uintptr(43), device)
		return "设备"
	})
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceGetUdid, func(uintptr) string { return "device-udid" })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceGetState, func(uintptr) string { return "paired" })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceGetConnections, func(uintptr) string { return connections })
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceGetScreencapMethods, func(uintptr) uint64 { return ^uint64(0) })
	before := liveNativeObjects.Load()
	result, err := FindIOSDevices()
	require.NoError(t, err)
	require.Equal(t, `["partial discovery"]`, result.Warnings)
	require.Equal(t, []*IOSDevice{{Name: "设备", UDID: "device-udid", State: "paired", Connections: connections, ScreencapMethod: ios.ScreencapAll}}, result.Devices)
	require.EqualValues(t, 1, destroyed.Load())
	require.Equal(t, before, liveNativeObjects.Load())
}

func TestToolkit_IOSDiscoveryNullableCallback(t *testing.T) {
	destroyed := stubIOSDiscovery(t)
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(connection int32, options string, callback native.MaaToolkitIOSDeviceCancelCallbackPtr, userdata, list uintptr) bool {
		require.Equal(t, int32(ios.ConnectionWiFi), connection)
		require.Equal(t, `{"timeout_ms":0,"server_port":17424}`, options)
		require.Zero(t, callback)
		require.Zero(t, userdata)
		require.Equal(t, uintptr(42), list)
		return true
	})
	result, err := FindIOSDevicesWithOptions(ios.ConnectionWiFi, `{"timeout_ms":0,"server_port":17424}`, nil)
	require.NoError(t, err)
	require.NotNil(t, result.Devices)
	require.Empty(t, result.Devices)
	require.EqualValues(t, 1, destroyed.Load())
}

func TestToolkit_IOSDiscoveryFailuresKeepWarningsAndRelease(t *testing.T) {
	for _, failure := range []string{"allocation", "discovery", "malformed options", "missing device", "callback panic"} {
		t.Run(failure, func(t *testing.T) {
			destroyed := stubIOSDiscovery(t)
			beforeObjects, beforeCallbacks := liveNativeObjects.Load(), iosDiscoveryCancellationCount()
			options := "{}"
			var cancelled func() bool
			switch failure {
			case "allocation":
				replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListCreate, func() uintptr { return 0 })
			case "discovery", "malformed options":
				if failure == "malformed options" {
					options = "{"
				}
				replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(_ int32, got string, _ native.MaaToolkitIOSDeviceCancelCallbackPtr, _, _ uintptr) bool {
					require.Equal(t, options, got)
					return false
				})
			case "missing device":
				replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListSize, func(uintptr) uint64 { return 1 })
				replaceNativeForTest(t, &native.MaaToolkitIOSDeviceListAt, func(uintptr, uint64) uintptr { return 0 })
			case "callback panic":
				cancelled = func() bool { panic("callback failed") }
				replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(_ int32, _ string, _ native.MaaToolkitIOSDeviceCancelCallbackPtr, token, _ uintptr) bool {
					result := make(chan uintptr, 1)
					go func() { result <- _MaaToolkitIOSDeviceCancelCallbackAgent(token) }()
					require.Equal(t, uintptr(1), <-result)
					return false
				})
			}
			result, err := FindIOSDevicesWithOptions(ios.ConnectionAuto, options, cancelled)
			require.Error(t, err)
			if failure == "allocation" {
				require.Nil(t, result)
				require.Zero(t, destroyed.Load())
			} else {
				require.Equal(t, `["partial discovery"]`, result.Warnings)
				require.ErrorContains(t, err, result.Warnings)
				require.EqualValues(t, 1, destroyed.Load())
			}
			if failure == "callback panic" {
				require.ErrorContains(t, err, "callback failed")
			}
			require.Equal(t, beforeObjects, liveNativeObjects.Load())
			require.Equal(t, beforeCallbacks, iosDiscoveryCancellationCount())
		})
	}
}

func TestToolkit_IOSDiscoveryCancellationLifecycle(t *testing.T) {
	destroyed := stubIOSDiscovery(t)
	beforeObjects, beforeCallbacks := liveNativeObjects.Load(), iosDiscoveryCancellationCount()
	var calls atomic.Int64
	var mu sync.Mutex
	tokens := make(map[uintptr]bool)
	replaceNativeForTest(t, &native.MaaToolkitIOSDeviceFindWithOptions, func(_ int32, _ string, callback native.MaaToolkitIOSDeviceCancelCallbackPtr, token, _ uintptr) bool {
		if uintptr(callback) != iosDeviceCancelCallback || token == 0 {
			t.Error("discovery must use the fixed callback and a nonzero integer token")
		}
		mu.Lock()
		if tokens[token] {
			t.Error("discovery reused a callback token")
		}
		tokens[token] = true
		mu.Unlock()
		worker := make(chan uintptr, 1)
		go func() { worker <- _MaaToolkitIOSDeviceCancelCallbackAgent(token) }()
		return <-worker == 0
	})
	cancelled := func() bool { calls.Add(1); return false }
	for range 64 {
		_, err := FindIOSDevicesWithOptions(ios.ConnectionAuto, "", cancelled)
		require.NoError(t, err)
	}
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			_, err := FindIOSDevicesWithOptions(ios.ConnectionAuto, "", cancelled)
			if err != nil {
				t.Errorf("concurrent discovery: %v", err)
			}
		})
	}
	workers.Wait()
	require.EqualValues(t, 80, calls.Load())
	require.EqualValues(t, 80, destroyed.Load())
	require.Len(t, tokens, 80)
	for token := range tokens {
		require.Equal(t, uintptr(1), _MaaToolkitIOSDeviceCancelCallbackAgent(token), "released tokens must not invoke user callbacks")
	}
	require.EqualValues(t, 80, calls.Load())
	require.Equal(t, beforeCallbacks, iosDiscoveryCancellationCount())
	require.Equal(t, beforeObjects, liveNativeObjects.Load())
}

func TestToolkit_IOSDiscoveryNativeCancellation(t *testing.T) {
	before := iosDiscoveryCancellationCount()
	var calls atomic.Int64
	result, err := FindIOSDevicesWithOptions(ios.ConnectionAuto, "{}", func() bool { calls.Add(1); return true })
	require.Error(t, err)
	require.NotNil(t, result)
	require.Contains(t, result.Warnings, "cancelled")
	require.Positive(t, calls.Load(), "the real native function must invoke the cancellation trampoline")
	require.Equal(t, before, iosDiscoveryCancellationCount())
}

func TestToolkit_IOSDiscoveryNativeInvalidOptions(t *testing.T) {
	result, err := FindIOSDevicesWithOptions(ios.ConnectionAuto, "{", nil)
	require.Error(t, err)
	require.NotNil(t, result)
	require.Contains(t, result.Warnings, "Invalid iOS discovery options")
}
