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

// FindWindow finds a win32 window by class name and window name.
// Return the number of windows found that match the criteria. To get the corresponding
// window handle, use GetWindow.
//
// This function finds the function by exact match. See also SearchWindow.
func FindWindow(className, windowName string) uint64 {
	cClassName := C.CString(className)
	cWindowName := C.CString(windowName)
	defer func() {
		C.free(unsafe.Pointer(cClassName))
		C.free(unsafe.Pointer(cWindowName))
	}()
	return uint64(C.MaaToolkitFindWindow(cClassName, cWindowName))
}

// SearchWindow regex search a win32 window by class name and window name.
// Return the number of windows found that match the criteria. To get the corresponding
// window handle, use GetWindow.
func SearchWindow(className, windowName string) uint64 {
	cClassName := C.CString(className)
	cWindowName := C.CString(windowName)
	defer func() {
		C.free(unsafe.Pointer(cClassName))
		C.free(unsafe.Pointer(cWindowName))
	}()
	return uint64(C.MaaToolkitSearchWindow(cClassName, cWindowName))
}

// ListWindows lists all windows.
// Return the number of windows found. To get the corresponding window handle, use GetWindow.
func ListWindows() uint64 {
	return uint64(C.MaaToolkitListWindows())
}

// GetWindow returns the window handle by index.
func GetWindow(index uint64) unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetWindow(C.uint64_t(index)))
}

// GetCursorWindow returns the window handle of the window under the cursor.
// This uses the WindowFromPoint() system API.
func GetCursorWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetCursorWindow())
}

// GetDesktopWindow returns the desktop window handle.
// This uses the GetDesktopWindow() system API.
func GetDesktopWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetDesktopWindow())
}

// GetForegroundWindow returns the foreground window handle.
// This uses the GetForegroundWindow() system API.
func GetForegroundWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetForegroundWindow())
}

// GetWindowClassName returns the window class name by window handle.
func GetWindowClassName(hwnd unsafe.Pointer) (string, bool) {
	buffer := maa.NewStringBuffer()
	defer buffer.Destroy()
	got := C.MaaToolkitGetWindowClassName(C.MaaWin32Hwnd(hwnd), C.MaaStringBufferHandle(buffer.Handle()))
	return buffer.Get(), got != 0
}

// GetWindowWindowName returns the window window name by window handle.
func GetWindowWindowName(hwnd unsafe.Pointer) (string, bool) {
	buffer := maa.NewStringBuffer()
	defer buffer.Destroy()
	got := C.MaaToolkitGetWindowWindowName(C.MaaWin32Hwnd(hwnd), C.MaaStringBufferHandle(buffer.Handle()))
	return buffer.Get(), got != 0
}
