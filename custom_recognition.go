package maa

import (
	"image"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
)

var (
	customRecognitionRunnerCallbackID          uint64
	customRecognitionRunnerCallbackAgents      = make(map[uint64]CustomRecognitionRunner)
	customRecognitionRunnerCallbackAgentsMutex sync.RWMutex
)

func registerCustomRecognition(recognizer CustomRecognitionRunner) uint64 {
	id := atomic.AddUint64(&customRecognitionRunnerCallbackID, 1)

	customRecognitionRunnerCallbackAgentsMutex.Lock()
	customRecognitionRunnerCallbackAgents[id] = recognizer
	customRecognitionRunnerCallbackAgentsMutex.Unlock()

	return id
}

func unregisterCustomRecognition(id uint64) bool {
	customRecognitionRunnerCallbackAgentsMutex.Lock()
	defer customRecognitionRunnerCallbackAgentsMutex.Unlock()

	if _, ok := customRecognitionRunnerCallbackAgents[id]; !ok {
		return false
	}
	delete(customRecognitionRunnerCallbackAgents, id)
	return true
}

type CustomRecognitionArg struct {
	TaskDetail             *TaskDetail
	CurrentTaskName        string
	CustomRecognitionName  string
	CustomRecognitionParam string
	Img                    image.Image
	Roi                    Rect
}

type CustomRecognitionResult struct {
	Box    Rect
	Detail string
}

type CustomRecognitionRunner interface {
	Run(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool)
}

// CustomRecognition is an alias for CustomRecognitionRunner for backward compatibility.
//
// Deprecated: Use CustomRecognitionRunner instead. This type will be removed in the future.
type CustomRecognition = CustomRecognitionRunner

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

	customRecognitionRunnerCallbackAgentsMutex.RLock()
	recognition, exists := customRecognitionRunnerCallbackAgents[id]
	customRecognitionRunnerCallbackAgentsMutex.RUnlock()

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
			CurrentTaskName:        cStringToString(currentTaskName),
			CustomRecognitionName:  cStringToString(customRecognitionName),
			CustomRecognitionParam: cStringToString(customRecognitionParam),
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
