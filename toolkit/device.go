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

// PostFindDevice posts a request to find all ADB devices.
func PostFindDevice() bool {
	return C.MaaToolkitPostFindDevice() != 0
}

// PostFindDeviceWithAdb posts a request to find all ADB devices with a given ADB path.
func PostFindDeviceWithAdb(adbPath string) bool {
	cAdbPath := C.CString(adbPath)
	defer C.free(unsafe.Pointer(cAdbPath))
	return C.MaaToolkitPostFindDeviceWithAdb(cAdbPath) != 0
}

// IsFindDeviceCompleted checks if the find device request is completed.
func IsFindDeviceCompleted() bool {
	return C.MaaToolkitIsFindDeviceCompleted() != 0
}

// WaitForFindDeviceToComplete waits for the find device request to complete.
// Return the number of devices found.
func WaitForFindDeviceToComplete() uint64 {
	return uint64(C.MaaToolkitWaitForFindDeviceToComplete())
}

// GetDeviceCount returns the number of devices found.
func GetDeviceCount() uint64 {
	return uint64(C.MaaToolkitGetDeviceCount())
}

// GetDeviceName returns the device name by index.
func GetDeviceName(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceName(C.uint64_t(index)))
}

// GetDeviceAdbPath returns the device ADB path by index.
func GetDeviceAdbPath(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceAdbPath(C.uint64_t(index)))
}

// GetDeviceAdbSerial returns the device ADB serial by index.
func GetDeviceAdbSerial(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceAdbSerial(C.uint64_t(index)))
}

// GetDeviceAdbControllerType returns the device ADB controller type by index.
func GetDeviceAdbControllerType(index uint64) maa.AdbControllerType {
	return maa.AdbControllerType(C.MaaToolkitGetDeviceAdbControllerType(C.uint64_t(index)))
}

// GetDeviceAdbConfig returns the device ADB config by index.
func GetDeviceAdbConfig(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceAdbConfig(C.uint64_t(index)))
}

type AdbDevice struct {
	Name           string
	AdbPath        string
	Address        string
	ControllerType maa.AdbControllerType
	Config         string
}

// AdbDevices returns the adb devices.
func AdbDevices() []AdbDevice {
	PostFindDevice()
	for !IsFindDeviceCompleted() {
	}
	count := GetDeviceCount()
	devices := make([]AdbDevice, count)
	var i uint64
	for i = 0; i < count; i++ {
		name := GetDeviceName(i)
		adbPath := GetDeviceAdbPath(i)
		address := GetDeviceAdbSerial(i)
		ctrlType := GetDeviceAdbControllerType(i)
		config := GetDeviceAdbConfig(i)
		devices[i] = AdbDevice{
			Name:           name,
			AdbPath:        adbPath,
			Address:        address,
			ControllerType: ctrlType,
			Config:         config,
		}
	}
	return devices
}
