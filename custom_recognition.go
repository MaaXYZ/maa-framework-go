package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "def.h"

extern uint8_t _MaaCustomRecognitionCallbackAgent(
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
	"image"
	"sync/atomic"
	"unsafe"
)

var (
	customRecognitionCallbackID     uint64
	customRecognitionCallbackAgents = make(map[uint64]CustomRecognition)
)

func registerCustomRecognition(recognizer CustomRecognition) uint64 {
	id := atomic.AddUint64(&customRecognitionCallbackID, 1)
	customRecognitionCallbackAgents[id] = recognizer
	return id
}

func unregisterCustomRecognition(id uint64) bool {
	if _, ok := customRecognitionCallbackAgents[id]; !ok {
		return false
	}
	delete(customRecognitionCallbackAgents, id)
	return true
}

type CustomRecognitionArg struct {
	TaskDetail             *TaskDetail
	CurrentTaskName        string
	CustomRecognizerName   string
	CustomRecognitionParam string
	Img                    image.Image
	Roi                    Rect
}

type CustomRecognition interface {
	Run(ctx *Context, arg *CustomRecognitionArg) (CustomRecognitionResult, bool)
}

type CustomRecognitionResult struct {
	Box    Rect
	Detail string
}

//export _MaaCustomRecognitionCallbackAgent
func _MaaCustomRecognitionCallbackAgent(
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
	recognizer := customRecognitionCallbackAgents[id]
	context := Context{handle: ctx}
	tasker := context.GetTasker()
	taskDetail := tasker.getTaskDetail(int64(taskId))
	imgBuffer := newImageBufferByHandle(unsafe.Pointer(img))
	imgImg := imgBuffer.Get()

	ret, ok := recognizer.Run(
		&Context{handle: ctx},
		&CustomRecognitionArg{
			TaskDetail:             taskDetail,
			CurrentTaskName:        C.GoString(currentTaskName),
			CustomRecognizerName:   C.GoString(customRecognizerName),
			CustomRecognitionParam: C.GoString(customRecognitionParam),
			Img:                    imgImg,
			Roi:                    newRectBufferByHandle(unsafe.Pointer(roi)).Get(),
		},
	)
	if ok {
		box := ret.Box
		outBoxRect := newRectBufferByHandle(unsafe.Pointer(outBox))
		outBoxRect.Set(box)
		outDetailString := newStringBufferByHandle(unsafe.Pointer(outDetail))
		outDetailString.Set(ret.Detail)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
