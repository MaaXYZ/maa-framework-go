package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"unsafe"
)

type SyncContext C.MaaSyncContextHandle

func RunTask(syncCtx SyncContext, taskName, param string) bool {
	cTaskName := C.CString(taskName)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cParam))
	}()
	return C.MaaSyncContextRunTask(syncCtx, cTaskName, cParam) != 0
}

func RunRecognition(syncCtx SyncContext, image ImageBuffer, taskName, taskParam string) (RectBuffer, StringBuffer, bool) {
	cTaskName := C.CString(taskName)
	cTaskParam := C.CString(taskParam)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cTaskParam))
	}()

	outBox := NewRect()
	outDetail := NewString()
	ret := C.MaaSyncContextRunRecognition(
		syncCtx,
		C.MaaImageBufferHandle(image.Handle()),
		cTaskName,
		cTaskParam,
		C.MaaRectHandle(outBox.Handle()),
		C.MaaStringBufferHandle(outDetail.Handle()),
	)
	return outBox, outDetail, ret != 0
}

func RunAction(syncCtx SyncContext, taskName, taskParam string, curBox RectBuffer, curRecDetail string) bool {
	cTaskName := C.CString(taskName)
	cTaskParam := C.CString(taskParam)
	cCurRecDetail := C.CString(curRecDetail)
	defer func() {
		C.free(unsafe.Pointer(cTaskName))
		C.free(unsafe.Pointer(cTaskParam))
		C.free(unsafe.Pointer(cCurRecDetail))
	}()
	return C.MaaSyncContextRunAction(syncCtx, cTaskName, cTaskParam, C.MaaRectHandle(curBox.Handle()), cCurRecDetail) != 0
}

func Click(syncCtx SyncContext, x, y int32) bool {
	return C.MaaSyncContextClick(syncCtx, C.int32_t(x), C.int32_t(y)) != 0
}

func Swipe(syncCtx SyncContext, x1, y1, x2, y2, duration int32) bool {
	return C.MaaSyncContextSwipe(syncCtx, C.int32_t(x1), C.int32_t(y1), C.int32_t(x2), C.int32_t(y2), C.int32_t(duration)) != 0
}

func PressKey(syncCtx SyncContext, keycode int32) bool {
	return C.MaaSyncContextPressKey(syncCtx, C.int32_t(keycode)) != 0
}

func InputText(syncCtx SyncContext, text string) bool {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	return C.MaaSyncContextInputText(syncCtx, cText) != 0
}

func TouchDown(syncCtx SyncContext, contact, x, y, pressure int32) bool {
	return C.MaaSyncContextTouchDown(syncCtx, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)) != 0
}

func TouchMove(syncCtx SyncContext, contact, x, y, pressure int32) bool {
	return C.MaaSyncContextTouchMove(syncCtx, C.int32_t(contact), C.int32_t(x), C.int32_t(y), C.int32_t(pressure)) != 0
}

func TouchUp(syncCtx SyncContext, contact int32) bool {
	return C.MaaSyncContextTouchUp(syncCtx, C.int32_t(contact)) != 0
}

func Screencap(syncCtx SyncContext) (ImageBuffer, bool) {
	outImage := NewImage()
	ret := C.MaaSyncContextScreencap(syncCtx, C.MaaImageBufferHandle(outImage.Handle()))
	return outImage, ret != 0
}

func CacheImage(syncCtx SyncContext) (ImageBuffer, bool) {
	outImage := NewImage()
	ret := C.MaaSyncContextCachedImage(syncCtx, C.MaaImageBufferHandle(outImage.Handle()))
	return outImage, ret != 0
}
