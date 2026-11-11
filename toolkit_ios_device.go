package maa

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/MaaXYZ/maa-framework-go/v4/controller/ios"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/ebitengine/purego"
)

// IOSDevice is an owned snapshot of an iOS discovery record. State is
// discovered, paired, or connected; it does not prove service health.
type IOSDevice struct {
	Name            string
	UDID            string
	State           string
	Connections     string // JSON array preserving all USB/Wi-Fi candidates.
	ScreencapMethod ios.ScreencapMethod
}

// IOSDeviceDiscoveryResult contains copied discovery data. Warnings is the
// native JSON array of diagnostics, available on success and failure.
type IOSDeviceDiscoveryResult struct {
	Devices  []*IOSDevice
	Warnings string
}

// FindIOSDevices takes a default USB/Wi-Fi snapshot without pairing devices,
// installing DDI, or starting media. An empty Devices slice with nil error
// means no devices were found. Available diagnostics are returned even on error.
func FindIOSDevices() (*IOSDeviceDiscoveryResult, error) {
	return findIOSDevices(ios.ConnectionAuto, "", nil, false)
}

// FindIOSDevicesWithOptions discovers devices using a transport and JSON object
// options. Empty optionsJSON uses defaults: timeout_ms 3000 and server_port
// from XDB_PORT or 17423. timeout_ms accepts 0..60000; USB enumeration has a
// separate one-second budget.
//
// cancelled may be nil. It may run on a native worker thread and must be
// thread-safe, return promptly, and never reenter MaaFramework. Returning true
// requests cancellation. A panic is recovered at the callback boundary and
// reported as an error after native work stops. No callback is retained after
// this function returns. A non-nil result preserves available warnings on error.
func FindIOSDevicesWithOptions(connection ios.ConnectionMethod, optionsJSON string, cancelled func() bool) (*IOSDeviceDiscoveryResult, error) {
	return findIOSDevices(connection, optionsJSON, cancelled, true)
}

func findIOSDevices(connection ios.ConnectionMethod, optionsJSON string, cancelled func() bool, withOptions bool) (*IOSDeviceDiscoveryResult, error) {
	if strings.IndexByte(optionsJSON, 0) >= 0 {
		return nil, errors.New("iOS discovery options must contain no NUL")
	}
	list := native.MaaToolkitIOSDeviceListCreate()
	if list == 0 {
		return nil, errors.New("failed to create iOS device list")
	}
	liveNativeObjects.Add(1)
	defer func() {
		native.MaaToolkitIOSDeviceListDestroy(list)
		liveNativeObjects.Add(-1)
	}()

	var callback native.MaaToolkitIOSDeviceCancelCallbackPtr
	var token uintptr
	var cancellation *iosDiscoveryCancellation
	if cancelled != nil {
		cancellation = &iosDiscoveryCancellation{cancelled: cancelled}
		token = registerIOSDiscoveryCancellation(cancellation)
		defer unregisterIOSDiscoveryCancellation(token)
		callback = native.MaaToolkitIOSDeviceCancelCallbackPtr(iosDeviceCancelCallback)
	}

	var ok bool
	if withOptions {
		ok = native.MaaToolkitIOSDeviceFindWithOptions(int32(connection), optionsJSON, callback, token, list)
	} else {
		ok = native.MaaToolkitIOSDeviceFind(list)
	}
	result := &IOSDeviceDiscoveryResult{
		Devices:  make([]*IOSDevice, 0),
		Warnings: native.MaaToolkitIOSDeviceListGetWarnings(list),
	}
	if cancellation != nil {
		if err := cancellation.failure(); err != nil {
			return result, fmt.Errorf("%w; discovery warnings: %s", err, result.Warnings)
		}
	}
	if !ok {
		return result, fmt.Errorf("failed to find iOS devices: %s", result.Warnings)
	}
	for i, size := uint64(0), native.MaaToolkitIOSDeviceListSize(list); i < size; i++ {
		device := native.MaaToolkitIOSDeviceListAt(list, i)
		if device == 0 {
			return result, fmt.Errorf("failed to get iOS device at index %d; discovery warnings: %s", i, result.Warnings)
		}
		result.Devices = append(result.Devices, &IOSDevice{
			Name:            native.MaaToolkitIOSDeviceGetName(device),
			UDID:            native.MaaToolkitIOSDeviceGetUdid(device),
			State:           native.MaaToolkitIOSDeviceGetState(device),
			Connections:     native.MaaToolkitIOSDeviceGetConnections(device),
			ScreencapMethod: ios.ScreencapMethod(native.MaaToolkitIOSDeviceGetScreencapMethods(device)),
		})
	}
	return result, nil
}

type iosDiscoveryCancellation struct {
	mu        sync.Mutex
	cancelled func() bool
	err       error
}

func (c *iosDiscoveryCancellation) check() (result uintptr) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return 1
	}
	defer func() {
		if value := recover(); value != nil {
			c.err = fmt.Errorf("iOS discovery cancellation callback panicked: %v", value)
			result = 1
		}
	}()
	if c.cancelled() {
		return 1
	}
	return 0
}

func (c *iosDiscoveryCancellation) failure() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

var iosDiscoveryCancellations = struct {
	sync.RWMutex
	next    uintptr
	entries map[uintptr]*iosDiscoveryCancellation
}{entries: make(map[uintptr]*iosDiscoveryCancellation)}

var iosDeviceCancelCallback = purego.NewCallback(_MaaToolkitIOSDeviceCancelCallbackAgent)

func registerIOSDiscoveryCancellation(cancellation *iosDiscoveryCancellation) uintptr {
	iosDiscoveryCancellations.Lock()
	defer iosDiscoveryCancellations.Unlock()
	iosDiscoveryCancellations.next++
	token := iosDiscoveryCancellations.next
	iosDiscoveryCancellations.entries[token] = cancellation
	return token
}

func unregisterIOSDiscoveryCancellation(token uintptr) {
	iosDiscoveryCancellations.Lock()
	defer iosDiscoveryCancellations.Unlock()
	delete(iosDiscoveryCancellations.entries, token)
}

func _MaaToolkitIOSDeviceCancelCallbackAgent(userdata uintptr) uintptr {
	iosDiscoveryCancellations.RLock()
	cancellation := iosDiscoveryCancellations.entries[userdata]
	iosDiscoveryCancellations.RUnlock()
	if cancellation == nil {
		return 1
	}
	return cancellation.check()
}
