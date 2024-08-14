package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_recognizer.h"

extern uint8_t _AnalyzeAgent(
			MaaSyncContextHandle sync_context,
            const MaaImageBufferHandle image,
            MaaStringView task_name,
            MaaStringView custom_recognition_param,
            MaaTransparentArg recognizer_arg,
           	MaaRectHandle out_box,
			MaaStringBufferHandle out_detail);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"sync/atomic"
	"unsafe"
)

var (
	customRecognizerID       uint64
	customRecognizerNameToID = make(map[string]uint64)
	customRecognizerAgents   = make(map[uint64]CustomRecognizer)
)

func registerCustomRecognizer(name string, recognizer CustomRecognizer) uint64 {
	id := atomic.AddUint64(&customRecognizerID, 1)
	customRecognizerNameToID[name] = id
	customRecognizerAgents[id] = recognizer
	return id
}

func unregisterCustomRecognizer(name string) bool {
	id, ok := customRecognizerNameToID[name]
	if !ok {
		return false
	}
	delete(customRecognizerNameToID, name)
	delete(customRecognizerAgents, id)
	return ok
}

func clearCustomRecognizer() {
	customRecognizerNameToID = make(map[string]uint64)
	customRecognizerAgents = make(map[uint64]CustomRecognizer)
}

// CustomRecognizer defines an interface for custom recognizer.
// Implementers of this interface must embed a CustomRecognizerHandler struct
// and provide an implementation for the Analyze method.
type CustomRecognizer interface {
	Analyze(syncCtx SyncContext, img image.Image, taskName, RecognitionParam string) (AnalyzeResult, bool)

	Handle() unsafe.Pointer
	Destroy()
}

type AnalyzeResult struct {
	Box    buffer.Rect
	Detail string
}

type CustomRecognizerHandler struct {
	handle C.MaaCustomRecognizerHandle
}

func NewCustomRecognizerHandler() CustomRecognizerHandler {
	return CustomRecognizerHandler{
		handle: C.MaaCustomRecognizerHandleCreate(C.AnalyzeCallback(C._AnalyzeAgent)),
	}
}

func (r CustomRecognizerHandler) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r CustomRecognizerHandler) Destroy() {
	C.MaaCustomRecognizerHandleDestroy(r.handle)
}

//export _AnalyzeAgent
func _AnalyzeAgent(
	ctx C.MaaSyncContextHandle,
	img C.MaaImageBufferHandle,
	taskName, customRecognitionParam C.MaaStringView,
	recognizerArg C.MaaTransparentArg,
	outBox C.MaaRectHandle,
	outDetail C.MaaStringBufferHandle,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(unsafe.Pointer(recognizerArg)))
	rec := customRecognizerAgents[id]
	imgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(img))
	imgImg, err := imgBuffer.GetByRawData()
	if err != nil {
		return C.uint8_t(0)
	}

	ret, ok := rec.Analyze(
		SyncContext{handle: ctx},
		imgImg,
		C.GoString(taskName),
		C.GoString(customRecognitionParam),
	)
	if ok {
		box := ret.Box
		outBoxRect := buffer.NewRectBufferByHandle(unsafe.Pointer(outBox))
		outBoxRect.Set(box)
		outDetailString := buffer.NewStringBufferByHandle(unsafe.Pointer(outDetail))
		outDetailString.Set(ret.Detail)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
