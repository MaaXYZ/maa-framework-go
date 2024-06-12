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

func PostFindDevice() bool {
	return C.MaaToolkitPostFindDevice() != 0
}

func PostFindDeviceWithAdb(adbPath string) bool {
	cAdbPath := C.CString(adbPath)
	defer C.free(unsafe.Pointer(cAdbPath))
	return C.MaaToolkitPostFindDeviceWithAdb(cAdbPath) != 0
}

func IsFindDeviceCompleted() bool {
	return C.MaaToolkitIsFindDeviceCompleted() != 0
}

func WaitForFindDeviceToComplete() uint64 {
	return uint64(C.MaaToolkitWaitForFindDeviceToComplete())
}

func GetDeviceCount() uint64 {
	return uint64(C.MaaToolkitGetDeviceCount())
}

func GetDeviceName(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceName(C.uint64_t(index)))
}

func GetDeviceAdbPath(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceAdbPath(C.uint64_t(index)))
}

func GetDeviceAdbSerial(index uint64) string {
	return C.GoString(C.MaaToolkitGetDeviceAdbSerial(C.uint64_t(index)))
}

func GetDeviceAdbControllerType(index uint64) maa.AdbControllerType {
	return maa.AdbControllerType(C.MaaToolkitGetDeviceAdbControllerType(C.uint64_t(index)))
}

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
