package maa

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>
*/
import "C"
import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/maa"
)

// AdbDevice represents a single ADB device with various properties about its information.
type AdbDevice struct {
	Name            string
	AdbPath         string
	Address         string
	ScreencapMethod AdbScreencapMethod
	InputMethod     AdbInputMethod
	Config          string
}

// DesktopWindow represents a single desktop window with various properties about its information.
type DesktopWindow struct {
	Handle     unsafe.Pointer
	ClassName  string
	WindowName string
}

type Toolkit struct{}

// NewToolkit creates a new toolkit instance.
func NewToolkit() *Toolkit {
	return &Toolkit{}
}

// ConfigInitOption inits the toolkit config option.
func (t *Toolkit) ConfigInitOption(userPath, defaultJson string) bool {
	cUserPath := C.CString(userPath)
	defer C.free(unsafe.Pointer(cUserPath))
	cDefaultJson := C.CString(defaultJson)
	defer C.free(unsafe.Pointer(cDefaultJson))

	return C.MaaToolkitConfigInitOption(cUserPath, cDefaultJson) != 0
}

// FindAdbDevices finds adb devices.
func (t *Toolkit) FindAdbDevices(specifiedAdb ...string) []*AdbDevice {
	listHandle := maa.MaaToolkitAdbDeviceListCreate()
	defer maa.MaaToolkitAdbDeviceListDestroy(listHandle)
	var got bool
	if len(specifiedAdb) > 0 {
		got = maa.MaaToolkitAdbDeviceFindSpecified(specifiedAdb[0], listHandle)
	} else {
		got = maa.MaaToolkitAdbDeviceFind(listHandle)
	}
	if !got {
		return nil
	}

	size := maa.MaaToolkitAdbDeviceListSize(listHandle)
	list := make([]*AdbDevice, size)
	for i := uint64(0); i < size; i++ {
		deviceHandle := maa.MaaToolkitAdbDeviceListAt(listHandle, i)
		name := maa.MaaToolkitAdbDeviceGetName(deviceHandle)
		adbPath := maa.MaaToolkitAdbDeviceGetAdbPath(deviceHandle)
		address := maa.MaaToolkitAdbDeviceGetAddress(deviceHandle)
		screencapMethod := AdbScreencapMethod(maa.MaaToolkitAdbDeviceGetScreencapMethods(deviceHandle))
		inputMethod := AdbInputMethod(maa.MaaToolkitAdbDeviceGetInputMethods(deviceHandle))
		config := maa.MaaToolkitAdbDeviceGetConfig(deviceHandle)
		list[i] = &AdbDevice{
			Name:            name,
			AdbPath:         adbPath,
			Address:         address,
			ScreencapMethod: screencapMethod,
			InputMethod:     inputMethod,
			Config:          config,
		}
	}
	return list
}

// FindDesktopWindows finds desktop windows.
func (t *Toolkit) FindDesktopWindows() []*DesktopWindow {
	listHandle := C.MaaToolkitDesktopWindowListCreate()
	defer C.MaaToolkitDesktopWindowListDestroy(listHandle)
	got := C.MaaToolkitDesktopWindowFindAll(listHandle)
	if got == 0 {
		return nil
	}

	size := uint64(C.MaaToolkitDesktopWindowListSize(listHandle))
	list := make([]*DesktopWindow, size)
	for i := uint64(0); i < size; i++ {
		windowHandle := C.MaaToolkitDesktopWindowListAt(listHandle, C.uint64_t(i))
		handle := unsafe.Pointer(C.MaaToolkitDesktopWindowGetHandle(windowHandle))
		className := C.GoString(C.MaaToolkitDesktopWindowGetClassName(windowHandle))
		windowName := C.GoString(C.MaaToolkitDesktopWindowGetWindowName(windowHandle))
		list[i] = &DesktopWindow{
			Handle:     handle,
			ClassName:  className,
			WindowName: windowName,
		}
	}
	return list
}

type piStoreValue struct {
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

var piStore = make(map[uint64]piStoreValue)

// RegisterPICustomRecognition registers a custom recognizer.
func (t *Toolkit) RegisterPICustomRecognition(instId uint64, name string, recognition CustomRecognition) {
	id := registerCustomRecognition(recognition)
	if _, ok := piStore[instId]; !ok {
		piStore[instId] = piStoreValue{
			CustomRecognizersCallbackID: make(map[string]uint64),
			CustomActionsCallbackID:     make(map[string]uint64),
		}
	}
	piStore[instId].CustomRecognizersCallbackID[name] = id

	maa.MaaToolkitProjectInterfaceRegisterCustomRecognition(
		instId,
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// RegisterPICustomAction registers a custom action.
func (t *Toolkit) RegisterPICustomAction(instId uint64, name string, action CustomAction) {
	id := registerCustomAction(action)
	if _, ok := piStore[instId]; !ok {
		piStore[instId] = piStoreValue{
			CustomRecognizersCallbackID: make(map[string]uint64),
			CustomActionsCallbackID:     make(map[string]uint64),
		}
	}
	piStore[instId].CustomActionsCallbackID[name] = id

	maa.MaaToolkitProjectInterfaceRegisterCustomAction(
		instId,
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// ClearPICustom unregisters all custom recognitions and actions for a given instance.
func (t *Toolkit) ClearPICustom(instId uint64) {
	value := piStore[instId]
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}
}

// RunCli runs the PI CLI.
func (t *Toolkit) RunCli(instId uint64, resourcePath, userPath string, directly bool, notify Notification) bool {
	id := registerNotificationCallback(notify)
	got := maa.MaaToolkitProjectInterfaceRunCli(
		instId,
		resourcePath,
		userPath,
		directly,
		MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got
}
