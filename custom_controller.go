package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_controller.h"
#include "def.h"

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);

extern uint8_t _ConnectAgent(void* handleArg);
extern uint8_t _RequestUUIDAgent(void* handle_arg, MaaStringBuffer* buffer);
extern uint8_t _StartAppAgent(const char* intent, void* handle_arg);
extern uint8_t _StopAppAgent(const char* intent, void* handle_arg);
extern uint8_t _ScreencapAgent(void* handle_arg, MaaImageBuffer* buffer);
extern uint8_t _ClickAgent(int32_t x, int32_t y, void* handle_arg);
extern uint8_t _SwipeAgent(
			int32_t x1,
			int32_t y1,
			int32_t x2,
			int32_t y2,
			int32_t duration,
			void* handle_arg);
extern uint8_t _TouchDownAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            void* handle_arg);
extern uint8_t _TouchMoveAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            void* handle_arg);
extern uint8_t _TouchUpAgent(int32_t contact, void* handle_arg);
extern uint8_t _PressKey(int32_t keycode, void* handle_arg);
extern uint8_t _InputText(const char* text, void* handle_arg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"image"
	"sync/atomic"
	"unsafe"
)

var (
	customControllerCallbacksID     uint64
	customControllerCallbacksAgents = make(map[uint64]CustomController)
)

func registerCustomControllerCallbacks(ctrl CustomController) uint64 {
	id := atomic.AddUint64(&customControllerCallbacksID, 1)
	customControllerCallbacksAgents[id] = ctrl
	return id
}

func unregisterCustomControllerCallbacks(id uint64) {
	delete(customControllerCallbacksAgents, id)
}

// CustomController defines an interface for custom controller.
// Implementers of this interface must embed a CustomControllerHandler struct
// and provide implementations for the following methods:
// Connect, RequestUUID, StartApp, StopApp,
// Screencap, Click, Swipe, TouchDown, TouchMove, TouchUp,
// PressKey and InputText.
type CustomController interface {
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
	handle *C.MaaCustomControllerCallbacks
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
func _ConnectAgent(handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.Connect()
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _RequestUUIDAgent
func _RequestUUIDAgent(handleArg unsafe.Pointer, uuidBuffer *C.MaaStringBuffer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	uuid, ok := ctrl.RequestUUID()
	if ok {
		uuidStrBuffer := buffer.NewStringBufferByHandle(unsafe.Pointer(uuidBuffer))
		uuidStrBuffer.Set(uuid)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StartAppAgent
func _StartAppAgent(intent C.StringView, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.StartApp(C.GoString(intent))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAppAgent
func _StopAppAgent(intent C.StringView, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.StopApp(C.GoString(intent))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _ScreencapAgent
func _ScreencapAgent(handleArg unsafe.Pointer, imgBuffer *C.MaaImageBuffer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	img, captured := ctrl.Screencap()
	if captured {
		imgImgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(imgBuffer))
		if ok := imgImgBuffer.SetRawData(img); ok {
			return C.uint8_t(1)
		}
	}
	return C.uint8_t(0)
}

//export _ClickAgent
func _ClickAgent(x, y C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.Click(int32(x), int32(y))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _SwipeAgent
func _SwipeAgent(x1, y1, x2, y2, duration C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.Swipe(int32(x1), int32(y1), int32(x2), int32(y2), int32(duration))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchDownAgent
func _TouchDownAgent(contact, x, y, pressure C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.TouchDown(int32(contact), int32(x), int32(y), int32(pressure))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchMoveAgent
func _TouchMoveAgent(contact, x, y, pressure C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.TouchMove(int32(contact), int32(x), int32(y), int32(pressure))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchUpAgent
func _TouchUpAgent(contact C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.TouchUp(int32(contact))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _PressKey
func _PressKey(key C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.PressKey(int32(key))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _InputText
func _InputText(text C.StringView, handleArg unsafe.Pointer) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	ok := ctrl.InputText(C.GoString(text))
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
