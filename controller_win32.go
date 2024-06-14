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

func NewWin32Controller(
	hWnd unsafe.Pointer,
	win32CtrlType Win32ControllerType,
	callback func(msg, detailsJson string),
) Controller {
	agent := &callbackAgent{callback: callback}
	handle := C.MaaWin32ControllerCreate(
		C.MaaWin32Hwnd(C.MaaWin32Hwnd(hWnd)),
		C.int32_t(win32CtrlType),
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(agent)),
	)
	return &controller{handle: handle}
}
