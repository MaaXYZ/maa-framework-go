package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "controller_custom.h"

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);

extern uint8_t _ConnectAgent(MaaTransparentArg handleArg);
extern uint8_t _RequestUUIDAgent(MaaTransparentArg handle_arg, MaaStringBufferHandle buffer);
extern uint8_t _StartAppAgent(_GoString_ intent, MaaTransparentArg handle_arg);
extern uint8_t _StopAppAgent(_GoString_ intent, MaaTransparentArg handle_arg);
extern uint8_t _ScreencapAgent(MaaTransparentArg handle_arg, MaaImageBufferHandle buffer);
extern uint8_t _ClickAgent(int32_t x, int32_t y, MaaTransparentArg handle_arg);
extern uint8_t _SwipeAgent(
			int32_t x1,
			int32_t y1,
			int32_t x2,
			int32_t y2,
			int32_t duration,
			MaaTransparentArg handle_arg);
extern uint8_t _TouchDownAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
extern uint8_t _TouchMoveAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
extern uint8_t _TouchUpAgent(int32_t contact, MaaTransparentArg handle_arg);
extern uint8_t _PressKey(int32_t keycode, MaaTransparentArg handle_arg);
extern uint8_t _InputText(MaaStringView text, MaaTransparentArg handle_arg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"unsafe"
)

// CustomControllerImpl defines an interface for custom controller.
// Implementers of this interface must embed a CustomControllerHandler struct
// and provide implementations for the following methods:
// Connect, RequestUUID, StartApp, StopApp,
// Screencap, Click, Swipe, TouchDown, TouchMove, TouchUp,
// PressKey and InputText.
type CustomControllerImpl interface {
	Connect() bool
	RequestUUID() (string, bool)
	StartApp(intent string) bool
	StopApp(intent string) bool
	Screencap() (image.Image, bool)
	Click(x, y int32) bool
	Swipe(x1, y1, x2, y2, duration int32) bool
	TouchDown(contact, x, y, pressure int32) bool
	TouchMove(contact, x, y, pressure int32) bool
	TouchUp(contact int32) bool
	PressKey(keycode int32) bool
	InputText(text string) bool

	Handle() unsafe.Pointer
	Destroy()
}

type CustomControllerHandler struct {
	handle C.MaaCustomControllerHandle
}

func NewCustomControllerHandler() CustomControllerHandler {
	return CustomControllerHandler{handle: C.MaaCustomControllerHandleCreate(
		C.ConnectCallback(C._ConnectAgent),
		C.RequestUUIDCallback(C._RequestUUIDAgent),
		C.StartAppCallback(C._StartAppAgent),
		C.StopAppCallback(C._StopAppAgent),
		C.ScreencapCallback(C._ScreencapAgent),
		C.ClickCallback(C._ClickAgent),
		C.SwipeCallback(C._SwipeAgent),
		C.TouchDownCallback(C._TouchDownAgent),
		C.TouchMoveCallback(C._TouchMoveAgent),
		C.TouchUpCallback(C._TouchUpAgent),
		C.PressKeyCallback(C._PressKey),
		C.InputTextCallback(C._InputText),
	)}
}

func (c CustomControllerHandler) Handle() unsafe.Pointer {
	return unsafe.Pointer(c.handle)
}

func (c CustomControllerHandler) Destroy() {
	C.MaaCustomControllerHandleDestroy(c.handle)
}

//export _ConnectAgent
func _ConnectAgent(handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(C.MaaTransparentArg(handleArg))
	ok := ctrl.Connect()
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _RequestUUIDAgent
func _RequestUUIDAgent(handleArg C.MaaTransparentArg, uuidBuffer C.MaaStringBufferHandle) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	uuid, ok := ctrl.RequestUUID()
	if ok {
		uuidStrBuffer := buffer.NewStringBufferByHandle(unsafe.Pointer(uuidBuffer))
		uuidStrBuffer.Set(uuid)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StartAppAgent
func _StartAppAgent(intent string, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.StartApp(intent)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAppAgent
func _StopAppAgent(intent string, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.StopApp(intent)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _ScreencapAgent
func _ScreencapAgent(handleArg C.MaaTransparentArg, imgBuffer C.MaaImageBufferHandle) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	img, ok := ctrl.Screencap()
	if ok {
		imgImgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(imgBuffer))
		err := imgImgBuffer.SetRawData(img)
		if err != nil {
			return C.uint8_t(0)
		}
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _ClickAgent
func _ClickAgent(x, y C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.Click(int32(x), int32(y))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _SwipeAgent
func _SwipeAgent(x1, y1, x2, y2, duration C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.Swipe(int32(x1), int32(y1), int32(x2), int32(y2), int32(duration))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchDownAgent
func _TouchDownAgent(contact, x, y, pressure C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.TouchDown(int32(contact), int32(x), int32(y), int32(pressure))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchMoveAgent
func _TouchMoveAgent(contact, x, y, pressure C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.TouchMove(int32(contact), int32(x), int32(y), int32(pressure))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchUpAgent
func _TouchUpAgent(contact C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.TouchUp(int32(contact))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _PressKey
func _PressKey(key C.int32_t, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.PressKey(int32(key))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _InputText
func _InputText(text C.MaaStringView, handleArg C.MaaTransparentArg) C.uint8_t {
	ctrl := *(*CustomControllerImpl)(unsafe.Pointer(handleArg))
	ok := ctrl.InputText(C.GoString(text))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

// NewCustomController creates a custom controller instance.
func NewCustomController(
	customCtrl CustomControllerImpl,
	callback func(msg, detailsJson string),
) Controller {
	id := registerCallback(callback)
	handle := C.MaaCustomControllerCreate(
		C.MaaCustomControllerHandle(customCtrl.Handle()),
		C.MaaTransparentArg(unsafe.Pointer(&customCtrl)),
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	)
	return &controller{handle: handle}
}
