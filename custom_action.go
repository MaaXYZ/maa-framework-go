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
	"sync/atomic"
	"unsafe"
)

var (
	customActionID       uint64
	customActionNameToID = make(map[string]uint64)
	customActionAgents   = make(map[uint64]CustomAction)
)

func registerCustomAction(name string, action CustomAction) uint64 {
	id := atomic.AddUint64(&customActionID, 1)
	customActionNameToID[name] = id
	customActionAgents[id] = action
	return id
}

// RegisterCustomAction is a temporary function that exposes the internal
// registerCustomAction functionality. This function is intended for internal
// use within the library and should not be used by external users. This function
// may be removed or changed in future versions without notice.
//
// DO NOT USE THIS FUNCTION IN YOUR CODE.
//
// This function is expected to be moved to an internal package in the next
// version of the library.
func RegisterCustomAction(name string, action CustomAction) uint64 {
	return registerCustomAction(name, action)
}

func unregisterCustomAction(name string) bool {
	id, ok := customActionNameToID[name]
	if !ok {
		return false
	}
	delete(customActionNameToID, name)
	delete(customActionAgents, id)
	return ok
}

// UnregisterCustomAction is a temporary function that exposes the internal
// registerCustomAction functionality. This function is intended for internal
// use within the library and should not be used by external users. This function
// may be removed or changed in future versions without notice.
//
// DO NOT USE THIS FUNCTION IN YOUR CODE.
//
// This function is expected to be moved to an internal package in the next
// version of the library.
func UnregisterCustomAction(name string) bool {
	return unregisterCustomAction(name)
}

func clearCustomAction() {
	customActionNameToID = make(map[string]uint64)
	customActionAgents = make(map[uint64]CustomAction)
}

type CustomAction interface {
	Run(ctx *Context, taskDetail *TaskDetail, currentTaskName, customActionName, customActionParam string, recognitionDetail *RecognitionDetail, box Rect) bool
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
	action := customActionAgents[id]
	context := &Context{handle: ctx}
	tasker := context.GetTasker()
	taskDetail := tasker.getTaskDetail(int64(taskId))
	recognitionDetail := tasker.getRecognitionDetail(int64(recId))
	curBoxRectBuffer := newRectBufferByHandle(unsafe.Pointer(box))

	ok := action.Run(
		&Context{handle: ctx},
		taskDetail,
		C.GoString(currentTaskName),
		C.GoString(customActionName),
		C.GoString(customActionParam),
		recognitionDetail,
		curBoxRectBuffer.Get(),
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
