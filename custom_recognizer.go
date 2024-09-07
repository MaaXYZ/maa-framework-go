package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

typedef struct MaaContext* MaaContextHandle;

typedef struct MaaRect* MaaRectHandle;

typedef struct MaaStringBuffer* MaaStringBufferHandle;

extern uint8_t _MaaCustomRecognizerCallback(
			MaaContextHandle ctx,
			int64_t task_id,
            const char* recognizer_name,
            const char* custom_recognition_param,
            const MaaImageBufferHandle image,
            void* recognizer_arg,
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
	customRecognizerCallbackID     uint64
	customRecognizerNameToID       = make(map[string]uint64)
	customRecognizerCallbackAgents = make(map[uint64]func(ctx *Context, taskId int64, recognizerName, customRecognitionParam string, img image.Image) (AnalyzeResult, bool))
)

func registerCustomRecognizer(name string, recognizerCallback func(ctx *Context, taskId int64, recognizerName, customRecognitionParam string, img image.Image) (AnalyzeResult, bool)) uint64 {
	id := atomic.AddUint64(&customRecognizerCallbackID, 1)
	customRecognizerNameToID[name] = id
	customRecognizerCallbackAgents[id] = recognizerCallback
	return id
}

func unregisterCustomRecognizer(name string) bool {
	id, ok := customRecognizerNameToID[name]
	if !ok {
		return false
	}
	delete(customRecognizerNameToID, name)
	delete(customRecognizerCallbackAgents, id)
	return ok
}

func clearCustomRecognizer() {
	customRecognizerNameToID = make(map[string]uint64)
	customRecognizerCallbackAgents = make(map[uint64]func(ctx *Context, taskId int64, recognizerName, customRecognitionParam string, img image.Image) (AnalyzeResult, bool))
}

type AnalyzeResult struct {
	Box    Rect
	Detail string
}

//export _MaaCustomRecognizerCallback
func _MaaCustomRecognizerCallback(
	ctx C.MaaContextHandle,
	taskId C.int64_t,
	recognizerName, customRecognitionParam C.CString,
	img C.MaaImageBufferHandle,
	recognizerArg unsafe.Pointer,
	outBox C.MaaRectHandle,
	outDetail C.MaaStringBufferHandle,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(recognizerArg))
	callback := customRecognizerCallbackAgents[id]
	imgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(img))
	imgImg, err := imgBuffer.GetByRawData()
	if err != nil {
		return C.uint8_t(0)
	}

	ret, ok := callback(
		&Context{handle: ctx},
		int64(taskId),
		C.GoString(recognizerName),
		C.GoString(customRecognitionParam),
		imgImg,
	)
	if ok {
		box := ret.Box
		outBoxRect := newRectBufferByHandle(unsafe.Pointer(outBox))
		outBoxRect.Set(box)
		outDetailString := buffer.NewStringBufferByHandle(unsafe.Pointer(outDetail))
		outDetailString.Set(ret.Detail)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
