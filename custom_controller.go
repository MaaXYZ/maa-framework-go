package maa

import (
	"image"
	"sync/atomic"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/internal/maa"
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
	handle uintptr
}

func NewCustomControllerHandler() CustomControllerHandler {
	return CustomControllerHandler{
		handle: maa.MaaCustomControllerCallbacksCreate(
			_ConnectAgent,
			_RequestUUIDAgent,
			_StartAppAgent,
			_StopAppAgent,
			_ScreencapAgent,
			_ClickAgent,
			_SwipeAgent,
			_TouchDownAgent,
			_TouchMoveAgent,
			_TouchUpAgent,
			_PressKey,
			_InputText,
		),
	}
}

func (c CustomControllerHandler) Handle() unsafe.Pointer {
	return unsafe.Pointer(c.handle)
}

func _ConnectAgent(handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.Connect()
}

func _RequestUUIDAgent(handleArg unsafe.Pointer, uuidBuffer uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	uuid, ok := ctrl.RequestUUID()
	if ok {
		uuidStrBuffer := buffer.NewStringBufferByHandle(uuidBuffer)
		uuidStrBuffer.Set(uuid)
		return true
	}
	return false
}

func _StartAppAgent(intent string, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.StartApp(intent)
}

func _StopAppAgent(intent string, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.StopApp(intent)
}

func _ScreencapAgent(handleArg unsafe.Pointer, imgBuffer uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	img, captured := ctrl.Screencap()
	if captured {
		imgImgBuffer := buffer.NewImageBufferByHandle(imgBuffer)
		if ok := imgImgBuffer.Set(img); ok {
			return true
		}
	}
	return false
}

func _ClickAgent(x, y int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.Click(x, y)
}

func _SwipeAgent(x1, y1, x2, y2, duration int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.Swipe(x1, y1, x2, y2, duration)
}

func _TouchDownAgent(contact, x, y, pressure int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.TouchDown(contact, x, y, pressure)
}

func _TouchMoveAgent(contact, x, y, pressure int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.TouchMove(contact, x, y, pressure)
}

func _TouchUpAgent(contact int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.TouchUp(contact)
}

func _PressKey(key int32, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.PressKey(key)
}

func _InputText(text string, handleArg unsafe.Pointer) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(handleArg))
	ctrl := customControllerCallbacksAgents[id]
	return ctrl.InputText(text)
}
