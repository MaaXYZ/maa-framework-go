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
	CtrlOptionScreenshotTargetLongSide

	// CtrlOptionScreenshotTargetShortSide Only one of long and short side can be set, and the other is automatically scaled according
	// to the aspect ratio.
	CtrlOptionScreenshotTargetShortSide

	// CtrlOptionDefaultAppPackageEntry For StartApp
	CtrlOptionDefaultAppPackageEntry

	// CtrlOptionDefaultAppPackage For StopApp
	CtrlOptionDefaultAppPackage

	// CtrlOptionRecording Dump all screenshots and actions
	//
	// Recording will evaluate to true if any of this or
	// MaaGlobalOptionEnum::MaaGlobalOption_Recording is true.
	CtrlOptionRecording
)

type Controller interface {
	Destroy()
	Handle() unsafe.Pointer

	SetScreenshotTargetLongSide(targetLongSide int) bool
	SetScreenshotTargetShortSide(targetShortSide int) bool
	SetDefaultAppPackageEntry(appPackage string) bool
	SetDefaultAppPackage(appPackage string) bool
	SetRecording(recording bool) bool

	PostConnect() Job
	PostClick(x, y int32) Job
	PostSwipe(x1, y1, x2, y2, duration int32) Job
	PostPressKey(keycode int32) Job
	PostInputText(text string) Job
	PostStartApp(intent string) Job
	PostStopApp(intent string) Job
	PostTouchDown(contact, x, y, pressure int32) Job
	PostTouchMove(contact, x, y, pressure int32) Job
	PostTouchUp(contact int32) Job
	PostScreencap() Job

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

func (c *controller) setOption(key CtrlOption, value unsafe.Pointer, valSize uintptr) bool {
	return C.MaaControllerSetOption(c.handle, C.int32_t(key), C.MaaOptionValue(value), C.uint64_t(valSize)) != 0
}

// SetScreenshotTargetLongSide sets screenshot target long side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1920
func (c *controller) SetScreenshotTargetLongSide(targetLongSide int) bool {
	targetLongSide32 := int32(targetLongSide)
	return c.setOption(
		CtrlOptionScreenshotTargetLongSide,
		unsafe.Pointer(&targetLongSide32),
		unsafe.Sizeof(targetLongSide32),
	)
}

// SetScreenshotTargetShortSide sets screenshot target short side.
// Only one of long and short side can be set, and the other is automatically scaled according to the aspect ratio.
//
// eg: 1080
func (c *controller) SetScreenshotTargetShortSide(targetShortSide int) bool {
	targetShortSide32 := int32(targetShortSide)
	return c.setOption(
		CtrlOptionScreenshotTargetShortSide,
		unsafe.Pointer(&targetShortSide32),
		unsafe.Sizeof(targetShortSide32),
	)
}

// SetDefaultAppPackageEntry sets app package for StartApp action.
//
// eg: "com.hypergryph.arknights/com.u8.sdk.U8UnityContext"
func (c *controller) SetDefaultAppPackageEntry(appPackage string) bool {
	cAppPackage := C.CString(appPackage)
	defer C.free(unsafe.Pointer(cAppPackage))

	return c.setOption(
		CtrlOptionDefaultAppPackageEntry,
		unsafe.Pointer(cAppPackage),
		uintptr(len(appPackage)),
	)
}

// SetDefaultAppPackage sets app package for StopApp action.
//
// eg: "com.hypergryph.arknights"
func (c *controller) SetDefaultAppPackage(appPackage string) bool {
	cAppPackage := C.CString(appPackage)
	defer C.free(unsafe.Pointer(cAppPackage))

	return c.setOption(
		CtrlOptionDefaultAppPackage,
		unsafe.Pointer(cAppPackage),
		uintptr(len(appPackage)),
	)
}

// SetRecording enables or disables the recording of all screenshots and actions.
func (c *controller) SetRecording(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}

	return c.setOption(
		CtrlOptionRecording,
		unsafe.Pointer(&cEnabled),
		unsafe.Sizeof(cEnabled),
	)
}

func (c *controller) PostConnect() Job {
	id := int64(C.MaaControllerPostConnection(c.handle))
	return NewJob(id, c.status)
}

func (c *controller) PostClick(x, y int32) Job {
	id := int64(C.MaaControllerPostClick(c.handle, C.int32_t(x), C.int32_t(y)))
	return NewJob(id, c.status)
}

func (c *controller) PostSwipe(x1, y1, x2, y2, duration int32) Job {
	id := int64(C.MaaControllerPostSwipe(c.handle, C.int32_t(x1), C.int32_t(y1), C.int32_t(x2), C.int32_t(y2), C.int32_t(duration)))
	return NewJob(id, c.status)
}

func (c *controller) PostPressKey(keycode int32) Job {
	id := int64(C.MaaControllerPostPressKey(c.handle, C.int32_t(keycode)))
	return NewJob(id, c.status)
}

func (c *controller) PostInputText(text string) Job {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	id := int64(C.MaaControllerPostInputText(c.handle, cText))
	return NewJob(id, c.status)
}

func (c *controller) PostStartApp(intent string) Job {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	id := int64(C.MaaControllerPostStartApp(c.handle, cIntent))
	return NewJob(id, c.status)
}

func (c *controller) PostStopApp(intent string) Job {
	cIntent := C.CString(intent)
	defer C.free(unsafe.Pointer(cIntent))
	id := int64(C.MaaControllerPostStopApp(c.handle, cIntent))
	return NewJob(id, c.status)
}

func (c *controller) PostTouchDown(contact, x, y, pressure int32) Job {
	id := int64(C.MaaControllerPostTouchDown(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
	return NewJob(id, c.status)
}

func (c *controller) PostTouchMove(contact, x, y, pressure int32) Job {
	id := int64(C.MaaControllerPostTouchMove(c.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)))
	return NewJob(id, c.status)
}

func (c *controller) PostTouchUp(contact int32) Job {
	id := int64(C.MaaControllerPostTouchUp(c.handle, C.int32_t(contact)))
	return NewJob(id, c.status)
}

func (c *controller) PostScreencap() Job {
	id := int64(C.MaaControllerPostScreencap(c.handle))
	return NewJob(id, c.status)
}

func (c *controller) status(id int64) Status {
	return Status(C.MaaControllerStatus(c.handle, C.int64_t(id)))
}

func (c *controller) Connected() bool {
	return C.MaaControllerConnected(c.handle) != 0
}

func (c *controller) GetImage() (ImageBuffer, bool) {
	image := NewImageBuffer()
	got := C.MaaControllerGetImage(c.handle, C.MaaImageBufferHandle(image.Handle()))
	return image, got != 0
}

func (c *controller) GetUUID() (string, bool) {
	uuid := NewStringBuffer()
	defer uuid.Destroy()
	got := C.MaaControllerGetUUID(c.handle, C.MaaStringBufferHandle(uuid.Handle()))
	if got == 0 {
		return "", false
	}
	return uuid.Get(), true
}
