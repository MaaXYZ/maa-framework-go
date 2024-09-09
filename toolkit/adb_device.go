package toolkit

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go"
	"unsafe"
)

// AdbDeviceFinder is a struct that helps in finding ADB devices.
type AdbDeviceFinder struct {
	handle *C.MaaToolkitAdbDeviceList
}

// NewAdbDeviceFinder creates a new AdbDeviceFinder instance.
func NewAdbDeviceFinder() *AdbDeviceFinder {
	return &AdbDeviceFinder{
		handle: C.MaaToolkitAdbDeviceListCreate(),
	}
}

// Destroy releases the AdbDeviceFinder.
func (f *AdbDeviceFinder) Destroy() {
	C.MaaToolkitAdbDeviceListDestroy(f.handle)
}

// Find posts a request to find all ADB devices.
func (f *AdbDeviceFinder) Find() bool {
	got := C.MaaToolkitAdbDeviceFind(f.handle)
	return got != 0
}

// FindSpecified posts a request to find all ADB devices with a given ADB path.
func (f *AdbDeviceFinder) FindSpecified(adbPath string) bool {
	cAdbPath := C.CString(adbPath)
	defer C.free(unsafe.Pointer(cAdbPath))

	got := C.MaaToolkitAdbDeviceFindSpecified(cAdbPath, f.handle)
	return got != 0
}

// size returns the number of devices found.
func (f *AdbDeviceFinder) size() uint64 {
	return uint64(C.MaaToolkitAdbDeviceListSize(f.handle))
}

// get returns AdbDevice by index.
func (f *AdbDeviceFinder) get(index uint64) *AdbDevice {
	handle := C.MaaToolkitAdbDeviceListAt(f.handle, C.uint64_t(index))
	if handle == nil {
		return nil
	}
	return &AdbDevice{
		handle: handle,
	}
}

// List returns a slice of all found ADB devices.
func (f *AdbDeviceFinder) List() []*AdbDevice {
	size := f.size()
	list := make([]*AdbDevice, size)
	for i := uint64(0); i < size; i++ {
		list[i] = f.get(i)
	}
	return list
}

// AdbDevice represents a single ADB device with various properties and methods to access its information.
type AdbDevice struct {
	handle *C.MaaToolkitAdbDevice
}

// GetName returns the device name.
func (d *AdbDevice) GetName() string {
	name := C.MaaToolkitAdbDeviceGetName(d.handle)
	return C.GoString(name)
}

// GetAdbPath returns the device ADB path.
func (d *AdbDevice) GetAdbPath() string {
	path := C.MaaToolkitAdbDeviceGetAdbPath(d.handle)
	return C.GoString(path)
}

// GetAddress returns the device ADB address.
func (d *AdbDevice) GetAddress() string {
	address := C.MaaToolkitAdbDeviceGetAddress(d.handle)
	return C.GoString(address)
}

// GetScreencapMethod returns the device adb screencap method.
func (d *AdbDevice) GetScreencapMethod() maa.AdbScreencapMethod {
	method := C.MaaToolkitAdbDeviceGetScreencapMethods(d.handle)
	return maa.AdbScreencapMethod(method)
}

// GetInputMethod returns the device adb input method.
func (d *AdbDevice) GetInputMethod() maa.AdbInputMethod {
	method := C.MaaToolkitAdbDeviceGetInputMethods(d.handle)
	return maa.AdbInputMethod(method)
}

// GetConfig returns the device ADB config.
func (d *AdbDevice) GetConfig() string {
	config := C.MaaToolkitAdbDeviceGetConfig(d.handle)
	return C.GoString(config)
}
