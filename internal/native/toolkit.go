package native

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var maaToolkit uintptr

var MaaToolkitConfigInitOption func(userPath, defaultJson string) bool

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

var (
	MaaToolkitDesktopWindowListCreate    func() uintptr
	MaaToolkitDesktopWindowListDestroy   func(handle uintptr)
	MaaToolkitDesktopWindowFindAll       func(buffer uintptr) bool
	MaaToolkitDesktopWindowListSize      func(list uintptr) uint64
	MaaToolkitDesktopWindowListAt        func(list uintptr, index uint64) uintptr
	MaaToolkitDesktopWindowGetHandle     func(window uintptr) unsafe.Pointer
	MaaToolkitDesktopWindowGetClassName  func(window uintptr) string
	MaaToolkitDesktopWindowGetWindowName func(window uintptr) string
)

func initToolkit(libDir string) error {
	libName := getMaaToolkitLibrary()
	libPath := filepath.Join(libDir, libName)

	handle, err := openLibrary(libPath)
	if err != nil {
		return &LibraryLoadError{
			LibraryName: "MaaToolkit",
			LibraryPath: libPath,
			Err:         err,
		}
	}

	maaToolkit = handle

	registerToolkit()

	return nil
}

func getMaaToolkitLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaToolkit.dylib"
	case "linux":
		return "libMaaToolkit.so"
	case "windows":
		return "MaaToolkit.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func registerToolkit() {
	// Config
	purego.RegisterLibFunc(&MaaToolkitConfigInitOption, maaToolkit, "MaaToolkitConfigInitOption")
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
	// DesktopWindow
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowListCreate, maaToolkit, "MaaToolkitDesktopWindowListCreate")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowListDestroy, maaToolkit, "MaaToolkitDesktopWindowListDestroy")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowFindAll, maaToolkit, "MaaToolkitDesktopWindowFindAll")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowListSize, maaToolkit, "MaaToolkitDesktopWindowListSize")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowListAt, maaToolkit, "MaaToolkitDesktopWindowListAt")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowGetHandle, maaToolkit, "MaaToolkitDesktopWindowGetHandle")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowGetClassName, maaToolkit, "MaaToolkitDesktopWindowGetClassName")
	purego.RegisterLibFunc(&MaaToolkitDesktopWindowGetWindowName, maaToolkit, "MaaToolkitDesktopWindowGetWindowName")
}

func unregisterToolkit() error {
	return unloadLibrary(maaToolkit)
}
