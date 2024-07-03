package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"unsafe"
)

type SyncContext struct {
	handle C.MaaSyncContextHandle
}

func (ctx SyncContext) RunTask(taskName, param string) bool {
	cTaskName := C.CString(taskName)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cParam))
	}()
	return C.MaaSyncContextRunTask(ctx.handle, cTaskName, cParam) != 0
}

type RecognitionResult struct {
	Box    Rect
	Detail string
}

func (ctx SyncContext) RunRecognition(image ImageBuffer, taskName, taskParam string) (RecognitionResult, bool) {
	cTaskName := C.CString(taskName)
	cTaskParam := C.CString(taskParam)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cTaskParam))
	}()

	outBox := NewRectBuffer()
	outDetail := NewStringBuffer()
	defer func() {
		outBox.Destroy()
		outDetail.Destroy()
	}()
	ret := C.MaaSyncContextRunRecognition(
		ctx.handle,
		C.MaaImageBufferHandle(image.Handle()),
		cTaskName,
		cTaskParam,
		C.MaaRectHandle(outBox.Handle()),
		C.MaaStringBufferHandle(outDetail.Handle()),
	)
	return RecognitionResult{
		Box:    outBox.Get(),
		Detail: outDetail.Get(),
	}, ret != 0
}

func (ctx SyncContext) RunAction(taskName, taskParam string, curBox Rect, curRecDetail string) bool {
	cTaskName := C.CString(taskName)
	cTaskParam := C.CString(taskParam)
	cCurRecDetail := C.CString(curRecDetail)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cTaskParam))
		C.free(unsafe.Pointer(cCurRecDetail))
	}()

	curBoxRectBuffer := NewRectBuffer()
	curBoxRectBuffer.Set(curBox)
	defer curBoxRectBuffer.Destroy()
	return C.MaaSyncContextRunAction(ctx.handle, cTaskName, cTaskParam, C.MaaRectHandle(curBoxRectBuffer.Handle()), cCurRecDetail) != 0
}

func (ctx SyncContext) Click(x, y int32) bool {
	return C.MaaSyncContextClick(ctx.handle, C.int32_t(x), C.int32_t(y)) != 0
}

func (ctx SyncContext) Swipe(x1, y1, x2, y2, duration int32) bool {
	return C.MaaSyncContextSwipe(ctx.handle, C.int32_t(x1), C.int32_t(y1), C.int32_t(x2), C.int32_t(y2), C.int32_t(duration)) != 0
}

func (ctx SyncContext) PressKey(keycode int32) bool {
	return C.MaaSyncContextPressKey(ctx.handle, C.int32_t(keycode)) != 0
}

func (ctx SyncContext) InputText(text string) bool {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return C.MaaSyncContextInputText(ctx.handle, cText) != 0
}

func (ctx SyncContext) TouchDown(contact, x, y, pressure int32) bool {
	return C.MaaSyncContextTouchDown(ctx.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)) != 0
}

func (ctx SyncContext) TouchMove(contact, x, y, pressure int32) bool {
	return C.MaaSyncContextTouchMove(ctx.handle, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)) != 0
}

func (ctx SyncContext) TouchUp(contact int32) bool {
	return C.MaaSyncContextTouchUp(ctx.handle, C.int32_t(contact)) != 0
}

func (ctx SyncContext) Screencap() (ImageBuffer, bool) {
	outImage := NewImageBuffer()
	ret := C.MaaSyncContextScreencap(ctx.handle, C.MaaImageBufferHandle(outImage.Handle()))
	return outImage, ret != 0
}

func (ctx SyncContext) CacheImage() (ImageBuffer, bool) {
	outImage := NewImageBuffer()
	ret := C.MaaSyncContextCachedImage(ctx.handle, C.MaaImageBufferHandle(outImage.Handle()))
	return outImage, ret != 0
}
