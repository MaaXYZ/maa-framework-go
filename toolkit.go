package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
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
	return native.MaaToolkitConfigInitOption(userPath, defaultJson)
}

// FindAdbDevices finds adb devices.
func FindAdbDevices(specifiedAdb ...string) []*AdbDevice {
	listHandle := native.MaaToolkitAdbDeviceListCreate()
	defer native.MaaToolkitAdbDeviceListDestroy(listHandle)
	var got bool
	if len(specifiedAdb) > 0 {
		got = native.MaaToolkitAdbDeviceFindSpecified(specifiedAdb[0], listHandle)
	} else {
		got = native.MaaToolkitAdbDeviceFind(listHandle)
	}
	if !got {
		return nil
	}

	size := native.MaaToolkitAdbDeviceListSize(listHandle)
	list := make([]*AdbDevice, size)
	for i := uint64(0); i < size; i++ {
		deviceHandle := native.MaaToolkitAdbDeviceListAt(listHandle, i)
		name := native.MaaToolkitAdbDeviceGetName(deviceHandle)
		adbPath := native.MaaToolkitAdbDeviceGetAdbPath(deviceHandle)
		address := native.MaaToolkitAdbDeviceGetAddress(deviceHandle)
		screencapMethod := AdbScreencapMethod(native.MaaToolkitAdbDeviceGetScreencapMethods(deviceHandle))
		inputMethod := AdbInputMethod(native.MaaToolkitAdbDeviceGetInputMethods(deviceHandle))
		config := native.MaaToolkitAdbDeviceGetConfig(deviceHandle)
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
	listHandle := native.MaaToolkitDesktopWindowListCreate()
	defer native.MaaToolkitDesktopWindowListDestroy(listHandle)
	got := native.MaaToolkitDesktopWindowFindAll(listHandle)
	if !got {
		return nil
	}

	size := native.MaaToolkitDesktopWindowListSize(listHandle)
	list := make([]*DesktopWindow, size)
	for i := uint64(0); i < size; i++ {
		windowHandle := native.MaaToolkitDesktopWindowListAt(listHandle, i)
		handle := native.MaaToolkitDesktopWindowGetHandle(windowHandle)
		className := native.MaaToolkitDesktopWindowGetClassName(windowHandle)
		windowName := native.MaaToolkitDesktopWindowGetWindowName(windowHandle)
		list[i] = &DesktopWindow{
			Handle:     handle,
			ClassName:  className,
			WindowName: windowName,
		}
	}
	return list
}
