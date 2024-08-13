package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"unsafe"
)

type Win32ControllerType int32

// Win32ControllerType
const (
	Win32ControllerTypeInvalid Win32ControllerType = iota

	Win32ControllerTypeTouchSendMessage
	Win32ControllerTypeTouchSeize

	Win32ControllerTypeKeySendMessage Win32ControllerType = 1 << 8
	Win32ControllerTypeKeySeize       Win32ControllerType = 2 << 8

	Win32ControllerTypeScreencapGDI            Win32ControllerType = 1 << 16
	Win32ControllerTypeScreencapDXGIDesktopDup Win32ControllerType = 2 << 16
	Win32ControllerTypeScreencapDXGIFramePool  Win32ControllerType = 4 << 16
)

// NewWin32Controller creates a win32 controller instance.
func NewWin32Controller(
	hWnd unsafe.Pointer,
	win32CtrlType Win32ControllerType,
	callback func(msg, detailsJson string),
) Controller {
	id := registerCallback(callback)
	handle := C.MaaWin32ControllerCreate(
		C.MaaWin32Hwnd(C.MaaWin32Hwnd(hWnd)),
		C.int32_t(win32CtrlType),
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	)
	return &controller{handle: handle}
}
