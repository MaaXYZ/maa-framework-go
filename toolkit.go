package maa

import (
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
