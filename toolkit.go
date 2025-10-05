package maa

import (
	"sync"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
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

// ConfigInitOption inits the toolkit config option.
func ConfigInitOption(userPath, defaultJson string) bool {
	return maa.MaaToolkitConfigInitOption(userPath, defaultJson)
}

// FindAdbDevices finds adb devices.
func FindAdbDevices(specifiedAdb ...string) []*AdbDevice {
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
func FindDesktopWindows() []*DesktopWindow {
	listHandle := maa.MaaToolkitDesktopWindowListCreate()
	defer maa.MaaToolkitDesktopWindowListDestroy(listHandle)
	got := maa.MaaToolkitDesktopWindowFindAll(listHandle)
	if !got {
		return nil
	}

	size := maa.MaaToolkitDesktopWindowListSize(listHandle)
	list := make([]*DesktopWindow, size)
	for i := uint64(0); i < size; i++ {
		windowHandle := maa.MaaToolkitDesktopWindowListAt(listHandle, i)
		handle := maa.MaaToolkitDesktopWindowGetHandle(windowHandle)
		className := maa.MaaToolkitDesktopWindowGetClassName(windowHandle)
		windowName := maa.MaaToolkitDesktopWindowGetWindowName(windowHandle)
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

var (
	piStore      = make(map[uint64]piStoreValue)
	piStoreMutex sync.RWMutex
)

// PIRegisterCustomRecognition registers a custom recognizer.
func PIRegisterCustomRecognition(instId uint64, name string, recognition CustomRecognition) {
	id := registerCustomRecognition(recognition)

	piStoreMutex.Lock()
	defer piStoreMutex.Unlock()

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

// PIRegisterCustomAction registers a custom action.
func PIRegisterCustomAction(instId uint64, name string, action CustomAction) {
	id := registerCustomAction(action)

	piStoreMutex.Lock()
	defer piStoreMutex.Unlock()

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

// PIClearCustom unregisters all custom recognitions and actions for a given instance.
func PIClearCustom(instId uint64) {
	piStoreMutex.Lock()
	defer piStoreMutex.Unlock()

	value, ok := piStore[instId]
	if !ok {
		return
	}

	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}
}

// PIRunCli runs the PI CLI.
func PIRunCli(instId uint64, resourcePath, userPath string, directly bool, notify Notification) bool {
	id := registerNotificationCallback(notify)
	got := maa.MaaToolkitProjectInterfaceRunCli(
		instId,
		resourcePath,
		userPath,
		directly,
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got
}
