package maa

import (
	"image"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/ebitengine/purego"
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

type ControllerFeature uint64

const (
	ControllerFeatureNone                               ControllerFeature = 0
	ControllerFeatureUseMouseDownAndUpInsteadOfClick    ControllerFeature = 1
	ControllerFeatureUseKeyboardDownAndUpInsteadOfClick ControllerFeature = 1 << 1
)

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
}

type MaaCustomControllerCallbacks struct {
	Connect     uintptr
	RequestUUID uintptr
	GetFeature  uintptr
	StartApp    uintptr
	StopApp     uintptr
	Screencap   uintptr
	Click       uintptr
	Swipe       uintptr
	TouchDown   uintptr
	TouchMove   uintptr
	TouchUp     uintptr
	ClickKey    uintptr
	InputText   uintptr
	KeyDown     uintptr
	KeyUp       uintptr
}

var customControllerCallbacksHandle = new(MaaCustomControllerCallbacks)

func init() {
	customControllerCallbacksHandle.Connect = purego.NewCallback(_ConnectAgent)
	customControllerCallbacksHandle.RequestUUID = purego.NewCallback(_RequestUUIDAgent)
	customControllerCallbacksHandle.GetFeature = purego.NewCallback(_GetFeatureAgent)
	customControllerCallbacksHandle.StartApp = purego.NewCallback(_StartAppAgent)
	customControllerCallbacksHandle.StopApp = purego.NewCallback(_StopAppAgent)
	customControllerCallbacksHandle.Screencap = purego.NewCallback(_ScreencapAgent)
	customControllerCallbacksHandle.Click = purego.NewCallback(_ClickAgent)
	customControllerCallbacksHandle.Swipe = purego.NewCallback(_SwipeAgent)
	customControllerCallbacksHandle.TouchDown = purego.NewCallback(_TouchDownAgent)
	customControllerCallbacksHandle.TouchMove = purego.NewCallback(_TouchMoveAgent)
	customControllerCallbacksHandle.TouchUp = purego.NewCallback(_TouchUpAgent)
	customControllerCallbacksHandle.ClickKey = purego.NewCallback(_ClickKey)
	customControllerCallbacksHandle.InputText = purego.NewCallback(_InputText)
	customControllerCallbacksHandle.KeyDown = purego.NewCallback(_KeyDown)
	customControllerCallbacksHandle.KeyUp = purego.NewCallback(_KeyUp)
}

func _ConnectAgent(handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.Connect() {
		return uintptr(1)
	}
	return uintptr(0)
}

func _RequestUUIDAgent(handleArg uintptr, uuidBuffer uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	uuid, ok := ctrl.RequestUUID()
	if ok {
		uuidStrBuffer := buffer.NewStringBufferByHandle(uuidBuffer)
		uuidStrBuffer.Set(uuid)
		return uintptr(1)
	}
	return uintptr(0)
}

func _GetFeatureAgent(handleArg uintptr) ControllerFeature {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return ControllerFeatureNone
	}

	return ctrl.GetFeature()
}

func _StartAppAgent(intent *byte, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.StartApp(cStringToString(intent)) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _StopAppAgent(intent *byte, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.StopApp(cStringToString(intent)) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _ScreencapAgent(handleArg uintptr, imgBuffer uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	img, captured := ctrl.Screencap()
	if captured {
		imgImgBuffer := buffer.NewImageBufferByHandle(imgBuffer)
		if ok := imgImgBuffer.Set(img); ok {
			return uintptr(1)
		}
	}
	return uintptr(0)
}

func _ClickAgent(x, y int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.Click(x, y) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _SwipeAgent(x1, y1, x2, y2, duration int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.Swipe(x1, y1, x2, y2, duration) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _TouchDownAgent(contact, x, y, pressure int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.TouchDown(contact, x, y, pressure) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _TouchMoveAgent(contact, x, y, pressure int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.TouchMove(contact, x, y, pressure) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _TouchUpAgent(contact int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.TouchUp(contact) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _ClickKey(key int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.ClickKey(key) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _InputText(text *byte, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.InputText(cStringToString(text)) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _KeyDown(keycode int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.KeyDown(keycode) {
		return uintptr(1)
	}
	return uintptr(0)
}

func _KeyUp(keycode int32, handleArg uintptr) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(handleArg)

	customControllerCallbacksAgentsMutex.RLock()
	ctrl, exists := customControllerCallbacksAgents[id]
	customControllerCallbacksAgentsMutex.RUnlock()

	if !exists || ctrl == nil {
		return uintptr(0)
	}

	if ctrl.KeyUp(keycode) {
		return uintptr(1)
	}
	return uintptr(0)
}
