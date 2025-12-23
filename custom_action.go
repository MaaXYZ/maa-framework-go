package maa

import (
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
)

var (
	customActionCallbackID          uint64
	customActionCallbackAgents      = make(map[uint64]CustomAction)
	customActionCallbackAgentsMutex sync.RWMutex
)

func registerCustomAction(action CustomAction) uint64 {
	id := atomic.AddUint64(&customActionCallbackID, 1)

	customActionCallbackAgentsMutex.Lock()
	customActionCallbackAgents[id] = action
	customActionCallbackAgentsMutex.Unlock()

	return id
}

func unregisterCustomAction(id uint64) bool {
	customActionCallbackAgentsMutex.Lock()
	defer customActionCallbackAgentsMutex.Unlock()

	if _, ok := customActionCallbackAgents[id]; !ok {
		return false
	}
	delete(customActionCallbackAgents, id)
	return true
}

type CustomActionArg struct {
	TaskDetail        *TaskDetail
	CurrentTaskName   string
	CustomActionName  string
	CustomActionParam string
	RecognitionDetail *RecognitionDetail
	Box               Rect
}

type CustomAction interface {
	Run(ctx *Context, arg *CustomActionArg) bool
}

func _MaaCustomActionCallbackAgent(
	context uintptr,
	taskId int64,
	currentTaskName, customActionName, customActionParam *byte,
	recoId int64,
	box uintptr,
	transArg uintptr,
) uint64 {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	customActionCallbackAgentsMutex.RLock()
	action, exists := customActionCallbackAgents[id]
	customActionCallbackAgentsMutex.RUnlock()

	if !exists || action == nil {
		return 0
	}

	ctx := &Context{handle: context}
	tasker := ctx.GetTasker()
	taskDetail := tasker.getTaskDetail(taskId)
	recognitionDetail := tasker.getRecognitionDetail(recoId)
	curBoxRectBuffer := buffer.NewRectBufferByHandle(box)

	ok := action.Run(
		&Context{handle: context},
		&CustomActionArg{
			TaskDetail:        taskDetail,
			CurrentTaskName:   cStringToString(currentTaskName),
			CustomActionName:  cStringToString(customActionName),
			CustomActionParam: cStringToString(customActionParam),
			RecognitionDetail: recognitionDetail,
			Box:               curBoxRectBuffer.Get(),
		},
	)
	if ok {
		return 1
	}
	return 0
}
