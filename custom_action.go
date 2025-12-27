package maa

import (
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
)

var (
	customActionRunnerCallbackID          uint64
	customActionRunnerCallbackAgents      = make(map[uint64]CustomActionRunner)
	customActionRunnerCallbackAgentsMutex sync.RWMutex
)

func registerCustomAction(action CustomActionRunner) uint64 {
	id := atomic.AddUint64(&customActionRunnerCallbackID, 1)

	customActionRunnerCallbackAgentsMutex.Lock()
	customActionRunnerCallbackAgents[id] = action
	customActionRunnerCallbackAgentsMutex.Unlock()

	return id
}

func unregisterCustomAction(id uint64) bool {
	customActionRunnerCallbackAgentsMutex.Lock()
	defer customActionRunnerCallbackAgentsMutex.Unlock()

	if _, ok := customActionRunnerCallbackAgents[id]; !ok {
		return false
	}
	delete(customActionRunnerCallbackAgents, id)
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

type CustomActionRunner interface {
	Run(ctx *Context, arg *CustomActionArg) bool
}

// CustomAction is an alias for CustomActionRunner for backward compatibility.
//
// Deprecated: Use CustomActionRunner instead. This type will be removed in the future.
type CustomAction = CustomActionRunner

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

	customActionRunnerCallbackAgentsMutex.RLock()
	action, exists := customActionRunnerCallbackAgents[id]
	customActionRunnerCallbackAgentsMutex.RUnlock()

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
