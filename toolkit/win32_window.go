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

func FindWindow(className, windowName string) uint64 {
	cClassName := C.CString(className)
	cWindowName := C.CString(windowName)
	defer func() {
		C.free(unsafe.Pointer(cClassName))
		C.free(unsafe.Pointer(cWindowName))
	}()
	return uint64(C.MaaToolkitFindWindow(cClassName, cWindowName))
}

func SearchWindow(className, windowName string) uint64 {
	cClassName := C.CString(className)
	cWindowName := C.CString(windowName)
	defer func() {
		C.free(unsafe.Pointer(cClassName))
		C.free(unsafe.Pointer(cWindowName))
	}()
	return uint64(C.MaaToolkitSearchWindow(cClassName, cWindowName))
}

func ListWindows() uint64 {
	return uint64(C.MaaToolkitListWindows())
}

func GetWindow(index uint64) unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetWindow(C.uint64_t(index)))
}

func GetCursorWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetCursorWindow())
}

func GetDesktopWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetDesktopWindow())
}

func GetForegroundWindow() unsafe.Pointer {
	return unsafe.Pointer(C.MaaToolkitGetForegroundWindow())
}

func GetWindowClassName(hwnd unsafe.Pointer) (string, bool) {
	buffer := maa.NewString()
	defer buffer.Destroy()
	got := C.MaaToolkitGetWindowClassName(C.MaaWin32Hwnd(hwnd), C.MaaStringBufferHandle(buffer.Handle()))
	return buffer.Get(), got != 0
}

func GetWindowWindowName(hwnd unsafe.Pointer) (string, bool) {
	buffer := maa.NewString()
	defer buffer.Destroy()
	got := C.MaaToolkitGetWindowWindowName(C.MaaWin32Hwnd(hwnd), C.MaaStringBufferHandle(buffer.Handle()))
	return buffer.Get(), got != 0
}
