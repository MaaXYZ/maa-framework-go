package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "def.h"

extern uint8_t _MaaCustomRecognizerCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_recognizer_name,
	const char* custom_recognition_param,
	const MaaImageBuffer* image,
	const MaaRect* roi,
	void* recognizer_arg,
	MaaRect* out_box,
	MaaStringBuffer* out_detail);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"image"
	"sync/atomic"
	"unsafe"
)

var (
	customRecognizerCallbackID     uint64
	customRecognizerNameToID       = make(map[string]uint64)
	customRecognizerCallbackAgents = make(map[uint64]CustomRecognizer)
)

func registerCustomRecognizer(name string, recognizer CustomRecognizer) uint64 {
	id := atomic.AddUint64(&customRecognizerCallbackID, 1)
	customRecognizerNameToID[name] = id
	customRecognizerCallbackAgents[id] = recognizer
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
	customRecognizerCallbackAgents = make(map[uint64]CustomRecognizer)
}

type CustomRecognizerArg struct {
	TaskDetail             *TaskDetail
	CurrentTaskName        string
	CustomRecognizerName   string
	CustomRecognitionParam string
	Img                    image.Image
	Roi                    Rect
}

type CustomRecognizer interface {
	Run(ctx *Context, arg *CustomRecognizerArg) (CustomRecognizerResult, bool)
}

type CustomRecognizerResult struct {
	Box    Rect
	Detail string
}

//export _MaaCustomRecognizerCallbackAgent
func _MaaCustomRecognizerCallbackAgent(
	ctx *C.MaaContext,
	taskId C.int64_t,
	currentTaskName, customRecognizerName, customRecognitionParam C.StringView,
	img C.ConstMaaImageBufferPtr,
	roi C.ConstMaaRectPtr,
	recognizerArg unsafe.Pointer,
	outBox *C.MaaRect,
	outDetail *C.MaaStringBuffer,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(recognizerArg))
	recognizer := customRecognizerCallbackAgents[id]
	context := Context{handle: ctx}
	tasker := context.GetTasker()
	taskDetail := tasker.getTaskDetail(int64(taskId))
	imgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(img))
	imgImg := imgBuffer.GetByRawData()

	ret, ok := recognizer.Run(
		&Context{handle: ctx},
		&CustomRecognizerArg{
			TaskDetail:             taskDetail,
			CurrentTaskName:        C.GoString(currentTaskName),
			CustomRecognizerName:   C.GoString(customRecognizerName),
			CustomRecognitionParam: C.GoString(customRecognitionParam),
			Img:                    imgImg,
			Roi:                    buffer.NewRectBufferByHandle(unsafe.Pointer(roi)).Get(),
		},
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
