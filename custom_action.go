package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "def.h"

extern uint8_t _MaaCustomActionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_action_name,
	const char* custom_action_param,
	int64_t rec_id,
	const MaaRect* box ,
	void* actionArg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"sync/atomic"
	"unsafe"
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

//export _MaaCustomActionCallbackAgent
func _MaaCustomActionCallbackAgent(
	ctx *C.MaaContext,
	taskId C.int64_t,
	currentTaskName, customActionName, customActionParam C.StringView,
	recId C.int64_t,
	box C.ConstMaaRectPtr,
	actionArg unsafe.Pointer,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(actionArg))
	action := customActionCallbackAgents[id]
	context := &Context{handle: ctx}
	tasker := context.GetTasker()
	taskDetail := tasker.getTaskDetail(int64(taskId))
	recognitionDetail := tasker.getRecognitionDetail(int64(recId))
	curBoxRectBuffer := buffer.NewRectBufferByHandle(unsafe.Pointer(box))

	ok := action.Run(
		&Context{handle: ctx},
		&CustomActionArg{
			TaskDetail:        taskDetail,
			CurrentTaskName:   C.GoString(currentTaskName),
			CustomActionName:  C.GoString(customActionName),
			CustomActionParam: C.GoString(customActionParam),
			RecognitionDetail: recognitionDetail,
			Box:               toMaaRect(curBoxRectBuffer.Get()),
		},
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
