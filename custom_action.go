package maa

import (
	"sync/atomic"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
)

var (
	customActionCallbackID     uint64
	customActionCallbackAgents = make(map[uint64]CustomAction)
)

func registerCustomAction(action CustomAction) uint64 {
	id := atomic.AddUint64(&customActionCallbackID, 1)
	customActionCallbackAgents[id] = action
	return id
}

func unregisterCustomAction(id uint64) bool {
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
	transArg unsafe.Pointer,
) uint64 {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(transArg))
	action := customActionCallbackAgents[id]
	ctx := &Context{handle: context}
	tasker := ctx.GetTasker()
	taskDetail := tasker.getTaskDetail(taskId)
	recognitionDetail := tasker.getRecognitionDetail(recoId)
	curBoxRectBuffer := buffer.NewRectBufferByHandle(unsafe.Pointer(box))

	ok := action.Run(
		&Context{handle: context},
		&CustomActionArg{
			TaskDetail:        taskDetail,
			CurrentTaskName:   bytePtrToString(currentTaskName),
			CustomActionName:  bytePtrToString(customActionName),
			CustomActionParam: bytePtrToString(customActionParam),
			RecognitionDetail: recognitionDetail,
			Box:               curBoxRectBuffer.Get(),
		},
	)
	if ok {
		return 1
	}
	return 0
}
