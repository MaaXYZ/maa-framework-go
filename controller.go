package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"unsafe"
)

type CtrlOption int32

// CtrlOption
const (
	CtrlOptionInvalid CtrlOption = iota

	// CtrlOptionScreenshotTargetLongSide Only one of long and short side can be set, and the other is automatically scaled according
	// to the aspect ratio.
	//
	// value: int, eg: 1920; val_size: sizeof(int)
	CtrlOptionScreenshotTargetLongSide

	// CtrlOptionScreenshotTargetShortSide Only one of long and short side can be set, and the other is automatically scaled according
	// to the aspect ratio.
	//
	// value: int, eg: 1080; val_size: sizeof(int)
	CtrlOptionScreenshotTargetShortSide

	// CtrlOptionDefaultAppPackageEntry For StartApp
	//
	// value: string, eg: "com.hypergryph.arknights/com.u8.sdk.U8UnityContext"; val_size: string length
	CtrlOptionDefaultAppPackageEntry

	// CtrlOptionDefaultAppPackage For StopApp
	//
	// value: string, eg: "com.hypergryph.arknights"; val_size: string length
	CtrlOptionDefaultAppPackage

	// CtrlOptionRecording Dump all screenshots and actions
	//
	// Recording will evaluate to true if any of this or
	// MaaGlobalOptionEnum::MaaGlobalOption_Recording is true.
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	CtrlOptionRecording
)

type Controller interface {
	Destroy()
	Handle() unsafe.Pointer
	SetOption(key CtrlOption, value unsafe.Pointer, valSize uint64) bool
	PostConnect() int64
	PostClick(x, y int32) int64
	PostSwipe(x1, y1, x2, y2, duration int32) int64
	PostPressKey(keycode int32) int64
	PostInputText(text string) int64
	PostStartApp(intent string) int64
	PostStopApp(intent string) int64
	PostTouchDown(contact, x, y, pressure int32) int64
	PostTouchMove(contact, x, y, pressure int32) int64
	PostTouchUp(contact int32) int64
	PostScreencap() int64
	Status(id int64) Status
	Wait(id int64) Status
	Connected() bool
	GetImage() (ImageBuffer, bool)
	GetUUID() (string, bool)
}

type controller struct {
	handle C.MaaControllerHandle
}

func (c *controller) Destroy() {
	C.MaaControllerDestroy(c.handle)
}

func (c *controller) Handle() unsafe.Pointer {
	return unsafe.Pointer(c.handle)
}

func (c *controller) SetOption(key CtrlOption, value unsafe.Pointer, valSize uint64) bool {
	return C.MaaControllerSetOption(c.handle, C.int32_t(key), C.MaaOptionValue(value), C.uint64_t(valSize)) != 0
}

func (c *controller) PostConnect() int64 {
	return int64(C.MaaControllerPostConnection(c.handle))
}

func (c *controller) PostClick(x, y int32) int64 {
	return int64(C.MaaControllerPostClick(c.handle, C.int32_t(x), C.int32_t(y)))
}

func (c *controller) PostSwipe(x1, y1, x2, y2, duration int32) int64 {
	return int64(C.MaaControllerPostSwipe(c.handle, C.int32_t(x1), C.int32_t(y1), C.int32_t(x2), C.int32_t(y2), C.int32_t(duration)))
}

func (c *controller) PostPressKey(keycode int32) int64 {
	return int64(C.MaaControllerPostPressKey(c.handle, C.int32_t(keycode)))
}

func (c *controller) PostInputText(text string) int64 {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return int64(C.MaaControllerPostInputText(c.handle, cText))
}

func (c *controller) PostStartApp(intent string) int64 {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	return int64(C.MaaControllerPostStartApp(c.handle, cIntent))
}

func (c *controller) PostStopApp(intent string) int64 {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	return int64(C.MaaControllerPostStopApp(c.handle, cIntent))
}

func (c *controller) PostTouchDown(contact, x, y, pressure int32) int64 {
	return int64(C.MaaControllerPostTouchDown(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
}

func (c *controller) PostTouchMove(contact, x, y, pressure int32) int64 {
	return int64(C.MaaControllerPostTouchMove(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
}

func (c *controller) PostTouchUp(contact int32) int64 {
	return int64(C.MaaControllerPostTouchUp(c.handle, C.int32_t(contact)))
}

func (c *controller) PostScreencap() int64 {
	return int64(C.MaaControllerPostScreencap(c.handle))
}

func (c *controller) Status(id int64) Status {
	return Status(C.MaaControllerStatus(c.handle, C.int64_t(id)))
}

func (c *controller) Wait(id int64) Status {
	return Status(C.MaaControllerWait(c.handle, C.int64_t(id)))
}

func (c *controller) Connected() bool {
	return C.MaaControllerConnected(c.handle) != 0
}

func (c *controller) GetImage() (ImageBuffer, bool) {
	image := NewImageBuffer()
	got := C.MaaControllerGetImage(c.handle, C.MaaImageBufferHandle(unsafe.Pointer(image.Handle())))
	return image, got != 0
}

func (c *controller) GetUUID() (string, bool) {
	uuid := NewStringBuffer()
	defer uuid.Destroy()
	got := C.MaaControllerGetUUID(c.handle, C.MaaStringBufferHandle(unsafe.Pointer(uuid.Handle())))
	if got == 0 {
		return "", false
	}
	return uuid.Get(), true
}
