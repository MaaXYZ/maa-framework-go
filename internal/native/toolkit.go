package native

import (
	"fmt"
	"runtime"
	"unsafe"
)

var maaToolkit uintptr

const maaToolkitName = "MaaToolkit"

// MaaToolkitConfigInitOption initializes the toolkit option file under
// userPath, creating it from the JSON object defaultJson when it does not
// exist yet, and applies the loaded options to the global framework
// configuration. It reports false when defaultJson is not valid option JSON
// or the file cannot be written or loaded.
var MaaToolkitConfigInitOption func(userPath, defaultJson string) bool

// MaaMacOSPermission defines the macOS permission type.
type MaaMacOSPermission int32

// MaaMacOSPermission values consumed by the MaaToolkitMacOS* permission
// functions: screen capture and accessibility.
const (
	MaaMacOSPermissionScreenCapture MaaMacOSPermission = 1
	MaaMacOSPermissionAccessibility MaaMacOSPermission = 2
)

// ADB device discovery functions. MaaToolkitAdbDeviceListCreate creates an
// empty device list and MaaToolkitAdbDeviceListDestroy frees it; the Find
// functions fill the list buffer with the discovered devices, scanning all
// known emulators plus the adb executable from PATH, or restricting
// discovery to the given adbPath. MaaToolkitAdbDeviceListSize returns the
// number of devices and MaaToolkitAdbDeviceListAt the device at index, a
// pointer borrowed from the list that stays valid until the list is
// destroyed or modified by another Find call. The Get accessors report that
// device's fields: a human-readable name, the adb
// path used to query it, the adb serial as address, the supported screencap
// and input method bitmasks, and the JSON config object that discovery
// generated for the device.
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
	MaaToolkitAdbDeviceGetScreencapMethods func(device uintptr) uint64
	MaaToolkitAdbDeviceGetInputMethods     func(device uintptr) uint64
	MaaToolkitAdbDeviceGetConfig           func(device uintptr) string
)

// Desktop window discovery functions. MaaToolkitDesktopWindowListCreate
// creates an empty window list and MaaToolkitDesktopWindowListDestroy frees
// it; MaaToolkitDesktopWindowFindAll fills the list buffer with the windows
// found, and MaaToolkitDesktopWindowListSize and
// MaaToolkitDesktopWindowListAt read the list contents. The Get accessors
// report a window's platform-specific fields: GetHandle returns the raw
// window handle (a Win32 HWND, or on macOS a CGWindowID force-cast to a
// pointer, never an NSWindow pointer), GetClassName the Win32 window class
// name or the macOS bundle identifier, and GetWindowName the window title.
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

// Gamescope instance discovery functions.
// MaaToolkitGamescopeInstanceListCreate creates an empty instance list and
// MaaToolkitGamescopeInstanceListDestroy frees it;
// MaaToolkitGamescopeInstanceFindAll fills the list buffer with the running
// gamescope instances ordered by display number, and
// MaaToolkitGamescopeInstanceListSize and
// MaaToolkitGamescopeInstanceListAt read the list contents. The Get
// accessors report an instance's display number, the PipeWire node ID of
// its capture stream, and the EIS socket path used to inject input events
// via libei; a zero node ID or an empty socket path means that endpoint is
// unavailable.
var (
	MaaToolkitGamescopeInstanceListCreate        func() uintptr
	MaaToolkitGamescopeInstanceListDestroy       func(handle uintptr)
	MaaToolkitGamescopeInstanceFindAll           func(buffer uintptr) bool
	MaaToolkitGamescopeInstanceListSize          func(list uintptr) uint64
	MaaToolkitGamescopeInstanceListAt            func(list uintptr, index uint64) uintptr
	MaaToolkitGamescopeInstanceGetDisplayNo      func(instance uintptr) uint32
	MaaToolkitGamescopeInstanceGetPipeWireNodeId func(instance uintptr) uint32
	MaaToolkitGamescopeInstanceGetEisSocketPath  func(instance uintptr) string
)

// xdg-desktop-portal ScreenCast session functions (Linux portal screencap).
// MaaToolkitPortalHelperCreate creates a helper and
// MaaToolkitPortalHelperDestroy closes the session and frees it, but does
// not close the PipeWire file descriptor handed out by the helper.
// MaaToolkitPortalHelperOpenStream runs the ScreenCast portal flow (create
// session, select sources, start stream) and connects to the PipeWire
// remote. GetPersist and SetPersist control whether the session can
// persist: with persist enabled the portal returns a restore token on
// start, which GetRestoreToken reports and SetRestoreToken accepts before
// OpenStream to resume a previous session. GetPipeWireFD returns the
// PipeWire socket file descriptor, negative while no stream is open, and
// GetPipeWireNodeID the stream node ID.
var (
	MaaToolkitPortalHelperCreate            func() uintptr
	MaaToolkitPortalHelperDestroy           func(helper uintptr)
	MaaToolkitPortalHelperOpenStream        func(helper uintptr) bool
	MaaToolkitPortalHelperGetPersist        func(helper uintptr) bool
	MaaToolkitPortalHelperSetPersist        func(helper uintptr, enable bool)
	MaaToolkitPortalHelperGetPipeWireFD     func(helper uintptr) int32
	MaaToolkitPortalHelperGetPipeWireNodeID func(helper uintptr) uint32
	MaaToolkitPortalHelperGetRestoreToken   func(helper uintptr) string
	MaaToolkitPortalHelperSetRestoreToken   func(helper uintptr, token string)
)

// macOS permission functions taking a MaaMacOSPermission value. They are
// only effective on macOS; on other platforms they log and return false.
// CheckPermission reports whether the permission is granted;
// RequestPermission shows the system prompt and reports whether the
// permission is granted now, without waiting for the user's response;
// RevealPermissionSettings opens the System Settings page for the
// permission.
var (
	MaaToolkitMacOSCheckPermission          func(perm MaaMacOSPermission) bool
	MaaToolkitMacOSRequestPermission        func(perm MaaMacOSPermission) bool
	MaaToolkitMacOSRevealPermissionSettings func(perm MaaMacOSPermission) bool
)

var toolkitEntries = []Entry{
	{&MaaToolkitConfigInitOption, "MaaToolkitConfigInitOption"},
	{&MaaToolkitAdbDeviceListCreate, "MaaToolkitAdbDeviceListCreate"},
	{&MaaToolkitAdbDeviceListDestroy, "MaaToolkitAdbDeviceListDestroy"},
	{&MaaToolkitAdbDeviceFind, "MaaToolkitAdbDeviceFind"},
	{&MaaToolkitAdbDeviceFindSpecified, "MaaToolkitAdbDeviceFindSpecified"},
	{&MaaToolkitAdbDeviceListSize, "MaaToolkitAdbDeviceListSize"},
	{&MaaToolkitAdbDeviceListAt, "MaaToolkitAdbDeviceListAt"},
	{&MaaToolkitAdbDeviceGetName, "MaaToolkitAdbDeviceGetName"},
	{&MaaToolkitAdbDeviceGetAdbPath, "MaaToolkitAdbDeviceGetAdbPath"},
	{&MaaToolkitAdbDeviceGetAddress, "MaaToolkitAdbDeviceGetAddress"},
	{&MaaToolkitAdbDeviceGetScreencapMethods, "MaaToolkitAdbDeviceGetScreencapMethods"},
	{&MaaToolkitAdbDeviceGetInputMethods, "MaaToolkitAdbDeviceGetInputMethods"},
	{&MaaToolkitAdbDeviceGetConfig, "MaaToolkitAdbDeviceGetConfig"},
	{&MaaToolkitDesktopWindowListCreate, "MaaToolkitDesktopWindowListCreate"},
	{&MaaToolkitDesktopWindowListDestroy, "MaaToolkitDesktopWindowListDestroy"},
	{&MaaToolkitDesktopWindowFindAll, "MaaToolkitDesktopWindowFindAll"},
	{&MaaToolkitDesktopWindowListSize, "MaaToolkitDesktopWindowListSize"},
	{&MaaToolkitDesktopWindowListAt, "MaaToolkitDesktopWindowListAt"},
	{&MaaToolkitDesktopWindowGetHandle, "MaaToolkitDesktopWindowGetHandle"},
	{&MaaToolkitDesktopWindowGetClassName, "MaaToolkitDesktopWindowGetClassName"},
	{&MaaToolkitDesktopWindowGetWindowName, "MaaToolkitDesktopWindowGetWindowName"},
	{&MaaToolkitGamescopeInstanceListCreate, "MaaToolkitGamescopeInstanceListCreate"},
	{&MaaToolkitGamescopeInstanceListDestroy, "MaaToolkitGamescopeInstanceListDestroy"},
	{&MaaToolkitGamescopeInstanceFindAll, "MaaToolkitGamescopeInstanceFindAll"},
	{&MaaToolkitGamescopeInstanceListSize, "MaaToolkitGamescopeInstanceListSize"},
	{&MaaToolkitGamescopeInstanceListAt, "MaaToolkitGamescopeInstanceListAt"},
	{&MaaToolkitGamescopeInstanceGetDisplayNo, "MaaToolkitGamescopeInstanceGetDisplayNo"},
	{&MaaToolkitGamescopeInstanceGetPipeWireNodeId, "MaaToolkitGamescopeInstanceGetPipeWireNodeId"},
	{&MaaToolkitGamescopeInstanceGetEisSocketPath, "MaaToolkitGamescopeInstanceGetEisSocketPath"},
	{&MaaToolkitPortalHelperCreate, "MaaToolkitPortalHelperCreate"},
	{&MaaToolkitPortalHelperDestroy, "MaaToolkitPortalHelperDestroy"},
	{&MaaToolkitPortalHelperOpenStream, "MaaToolkitPortalHelperOpenStream"},
	{&MaaToolkitPortalHelperGetPersist, "MaaToolkitPortalHelperGetPersist"},
	{&MaaToolkitPortalHelperSetPersist, "MaaToolkitPortalHelperSetPersist"},
	{&MaaToolkitPortalHelperGetPipeWireFD, "MaaToolkitPortalHelperGetPipeWireFD"},
	{&MaaToolkitPortalHelperGetPipeWireNodeID, "MaaToolkitPortalHelperGetPipeWireNodeID"},
	{&MaaToolkitPortalHelperGetRestoreToken, "MaaToolkitPortalHelperGetRestoreToken"},
	{&MaaToolkitPortalHelperSetRestoreToken, "MaaToolkitPortalHelperSetRestoreToken"},
	{&MaaToolkitMacOSCheckPermission, "MaaToolkitMacOSCheckPermission"},
	{&MaaToolkitMacOSRequestPermission, "MaaToolkitMacOSRequestPermission"},
	{&MaaToolkitMacOSRevealPermissionSettings, "MaaToolkitMacOSRevealPermissionSettings"},
}

func getMaaToolkitLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaToolkit.dylib"
	case "linux", "android":
		return "libMaaToolkit.so"
	case "windows":
		return "MaaToolkit.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}
