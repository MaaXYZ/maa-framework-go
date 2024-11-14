package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "def.h"
*/
import "C"
import (
	"image"
	"sync/atomic"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
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

func _MaaCustomRecognitionCallbackAgent(
	context uintptr,
	taskId int64,
	currentTaskName, customRecognitionName, customRecognitionParam *byte,
	image, roi uintptr,
	transArg unsafe.Pointer,
	outBox, outDetail uintptr,
) uint64 {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(transArg))
	recognizer := customRecognitionCallbackAgents[id]
	ctx := Context{handle: (*C.MaaContext)(unsafe.Pointer(context))}
	tasker := ctx.GetTasker()
	taskDetail := tasker.getTaskDetail(int64(taskId))
	imgBuffer := buffer.NewImageBufferByHandle(unsafe.Pointer(image))
	imgImg := imgBuffer.Get()

	ret, ok := recognizer.Run(
		&Context{handle: (*C.MaaContext)(unsafe.Pointer(context))},
		&CustomRecognitionArg{
			TaskDetail:             taskDetail,
			CurrentTaskName:        bytePtrToString(currentTaskName),
			CustomRecognizerName:   bytePtrToString(customRecognitionName),
			CustomRecognitionParam: bytePtrToString(customRecognitionParam),
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
		return 1
	}
	return 0
}
