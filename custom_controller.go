package maa

import (
	"image"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

var (
	customControllerCallbacksID          uint64
	customControllerCallbacksAgents      = make(map[uint64]CustomController)
	customControllerCallbacksAgentsMutex sync.RWMutex
)

func registerCustomControllerCallbacks(ctrl CustomController) uint64 {
	id := atomic.AddUint64(&customControllerCallbacksID, 1)

	customControllerCallbacksAgentsMutex.Lock()
	customControllerCallbacksAgents[id] = ctrl
	customControllerCallbacksAgentsMutex.Unlock()

	return id
}

func unregisterCustomControllerCallbacks(id uint64) {
	customControllerCallbacksAgentsMutex.Lock()
	delete(customControllerCallbacksAgents, id)
	customControllerCallbacksAgentsMutex.Unlock()
}

// CustomController defines an interface for custom controller.
// Implementers of this interface must embed a CustomControllerHandler struct
// and provide implementations for the following methods:
// Connect, RequestUUID, StartApp, StopApp,
// Screencap, Click, Swipe, TouchDown, TouchMove, TouchUp,
// ClickKey, InputText, KeyDown and KeyUp.
type CustomController interface {
	Connect() bool
	RequestUUID() (string, bool)
	GetFeature() ControllerFeature
	StartApp(intent string) bool
	StopApp(intent string) bool
	Screencap() (image.Image, bool)
	Click(x, y int32) bool
	Swipe(x1, y1, x2, y2, duration int32) bool
	TouchDown(contact, x, y, pressure int32) bool
	TouchMove(contact, x, y, pressure int32) bool
	TouchUp(contact int32) bool
	ClickKey(keycode int32) bool
	InputText(text string) bool
	KeyDown(keycode int32) bool
	KeyUp(keycode int32) bool

	Handle() uintptr
}

type CustomControllerHandler struct {
	handle uintptr
}

func NewCustomControllerHandler() CustomControllerHandler {
	return CustomControllerHandler{
		handle: maa.MaaCustomControllerCallbacksCreate(
			_ConnectAgent,
			_RequestUUIDAgent,
			_GetFeatureAgent,
			_StartAppAgent,
			_StopAppAgent,
			_ScreencapAgent,
			_ClickAgent,
			_SwipeAgent,
			_TouchDownAgent,
			_TouchMoveAgent,
			_TouchUpAgent,
			_ClickKey,
			_InputText,
			_KeyDown,
			_KeyUp,
		),
	}
}

func (c CustomControllerHandler) Handle() uintptr {
	return c.handle
}

func _ConnectAgent(handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.Connect()
}

func _RequestUUIDAgent(handleArg uintptr, uuidBuffer uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	uuid, ok := ctrl.RequestUUID()
	if ok {
		uuidStrBuffer := buffer.NewStringBufferByHandle(uuidBuffer)
		uuidStrBuffer.Set(uuid)
		return true
	}
	return false
}

type ControllerFeature = maa.MaaControllerFeature

const (
	MaaControllerFeatureNone                               = maa.MaaControllerFeature_None
	MaaControllerFeatureUseMouseDownAndUpInsteadOfClick    = maa.MaaControllerFeature_UseMouseDownAndUpInsteadOfClick
	MaaControllerFeatureUseKeyboardDownAndUpInsteadOfClick = maa.MaaControllerFeature_UseKeyboardDownAndUpInsteadOfClick
)

func _GetFeatureAgent(handleArg uintptr) ControllerFeature {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return MaaControllerFeatureNone
	}

	return ctrl.GetFeature()
}

func _StartAppAgent(intent string, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.StartApp(intent)
}

func _StopAppAgent(intent string, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.StopApp(intent)
}

func _ScreencapAgent(handleArg uintptr, imgBuffer uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	img, captured := ctrl.Screencap()
	if captured {
		imgImgBuffer := buffer.NewImageBufferByHandle(imgBuffer)
		if ok := imgImgBuffer.Set(img); ok {
			return true
		}
	}
	return false
}

func _ClickAgent(x, y int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.Click(x, y)
}

func _SwipeAgent(x1, y1, x2, y2, duration int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.Swipe(x1, y1, x2, y2, duration)
}

func _TouchDownAgent(contact, x, y, pressure int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.TouchDown(contact, x, y, pressure)
}

func _TouchMoveAgent(contact, x, y, pressure int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.TouchMove(contact, x, y, pressure)
}

func _TouchUpAgent(contact int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.TouchUp(contact)
}

func _ClickKey(key int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.ClickKey(key)
}

func _InputText(text string, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.InputText(text)
}

func _KeyDown(keycode int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.KeyDown(keycode)
}

func _KeyUp(keycode int32, handleArg uintptr) bool {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return false
	}

	return ctrl.KeyUp(keycode)
}
