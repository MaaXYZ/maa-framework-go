package maa

import (
	"image"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
)

var (
	customRecognitionCallbackID          uint64
	customRecognitionCallbackAgents      = make(map[uint64]CustomRecognition)
	customRecognitionCallbackAgentsMutex sync.RWMutex
)

func registerCustomRecognition(recognizer CustomRecognition) uint64 {
	id := atomic.AddUint64(&customRecognitionCallbackID, 1)

	customRecognitionCallbackAgentsMutex.Lock()
	customRecognitionCallbackAgents[id] = recognizer
	customRecognitionCallbackAgentsMutex.Unlock()

	return id
}

func unregisterCustomRecognition(id uint64) bool {
	customRecognitionCallbackAgentsMutex.Lock()
	defer customRecognitionCallbackAgentsMutex.Unlock()

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
	Run(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool)
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
	transArg uintptr,
	outBox, outDetail uintptr,
) uint64 {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	customRecognitionCallbackAgentsMutex.RLock()
	recognition, exists := customRecognitionCallbackAgents[id]
	customRecognitionCallbackAgentsMutex.RUnlock()

	if !exists || recognition == nil {
		return 0
	}

	ctx := Context{handle: context}
	tasker := ctx.GetTasker()
	taskDetail := tasker.getTaskDetail(taskId)
	imgBuffer := buffer.NewImageBufferByHandle(image)
	imgImg := imgBuffer.Get()

	ret, ok := recognition.Run(
		&Context{handle: context},
		&CustomRecognitionArg{
			TaskDetail:             taskDetail,
			CurrentTaskName:        bytePtrToString(currentTaskName),
			CustomRecognizerName:   bytePtrToString(customRecognitionName),
			CustomRecognitionParam: bytePtrToString(customRecognitionParam),
			Img:                    imgImg,
			Roi:                    buffer.NewRectBufferByHandle(roi).Get(),
		},
	)
	if ok {
		box := ret.Box
		outBoxRect := buffer.NewRectBufferByHandle(outBox)
		outBoxRect.Set(box)
		outDetailString := buffer.NewStringBufferByHandle(outDetail)
		outDetailString.Set(ret.Detail)
		return 1
	}
	return 0
}
