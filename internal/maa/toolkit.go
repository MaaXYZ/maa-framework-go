package maa

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	MaaToolkitProjectInterfaceRegisterCustomRecognition func(instId uint64, name string, recognition MaaCustomRecognitionCallback, transArg unsafe.Pointer)
	MaaToolkitProjectInterfaceRegisterCustomAction      func(instId uint64, name string, action MaaCustomActionCallback, transArg unsafe.Pointer)
	MaaToolkitProjectInterfaceRunCli                    func(instId uint64, resourcePath, userPath string, directly bool, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) bool
)

var (
	MaaToolkitAdbDeviceListCreate          func() uintptr
	MaaToolkitAdbDeviceListDestroy         func(handle uintptr)
	MaaToolkitAdbDeviceFind                func(buffer uintptr) bool
	MaaToolkitAdbDeviceFindSpecified       func(adbPath string, buffer uintptr) bool
	MaaToolkitAdbDeviceListSize            func(list uintptr) uint64
	MaaToolkitAdbDeviceListAt              func(list uintptr, index uint64) uintptr
	MaaToolkitAdbDeviceGetName             func(device uintptr) string
	MaaToolkitAdbDeviceGetAdbPath          func(device uintptr) string
	MaaToolkitAdbDeviceGetAddress          func(device uintptr) string
	MaaToolkitAdbDeviceGetScreencapMethods func(device uintptr) int32
	MaaToolkitAdbDeviceGetInputMethods     func(device uintptr) int32
	MaaToolkitAdbDeviceGetConfig           func(device uintptr) string
)

func init() {
	maaToolkit, err := openLibrary(getMaaToolkitLibrary())
	if err != nil {
		panic(err)
	}
	// ProjectInterface
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRegisterCustomRecognition, maaToolkit, "MaaToolkitProjectInterfaceRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRegisterCustomAction, maaToolkit, "MaaToolkitProjectInterfaceRegisterCustomAction")
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRunCli, maaToolkit, "MaaToolkitProjectInterfaceRunCli")
	// AdbDevice
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceListCreate, maaToolkit, "MaaToolkitAdbDeviceListCreate")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceListDestroy, maaToolkit, "MaaToolkitAdbDeviceListDestroy")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceFind, maaToolkit, "MaaToolkitAdbDeviceFind")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceFindSpecified, maaToolkit, "MaaToolkitAdbDeviceFindSpecified")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceListSize, maaToolkit, "MaaToolkitAdbDeviceListSize")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceListAt, maaToolkit, "MaaToolkitAdbDeviceListAt")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetName, maaToolkit, "MaaToolkitAdbDeviceGetName")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetAdbPath, maaToolkit, "MaaToolkitAdbDeviceGetAdbPath")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetAddress, maaToolkit, "MaaToolkitAdbDeviceGetAddress")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetScreencapMethods, maaToolkit, "MaaToolkitAdbDeviceGetScreencapMethods")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetInputMethods, maaToolkit, "MaaToolkitAdbDeviceGetInputMethods")
	purego.RegisterLibFunc(&MaaToolkitAdbDeviceGetConfig, maaToolkit, "MaaToolkitAdbDeviceGetConfig")
}
