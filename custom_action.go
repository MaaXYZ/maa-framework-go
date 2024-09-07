package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "def.h"

extern uint8_t _MaaCustomActionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char*  task_name,
	const char*  customActionParam,
	const MaaRect* box ,
	const char* recognition_detail,
	void* actionArg);
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

var (
	customActionCallbackID     uint64
	customActionNameToID       = make(map[string]uint64)
	customActionCallbackAgents = make(map[uint64]func(ctx *Context, taskId int64, actionName, customActionParam string, box Rect, recognitionDetail string) bool)
)

func registerCustomAction(name string, action func(ctx *Context, taskId int64, actionName, customActionParam string, box Rect, recognitionDetail string) bool) uint64 {
	id := atomic.AddUint64(&customActionCallbackID, 1)
	customActionNameToID[name] = id
	customActionCallbackAgents[id] = action
	return id
}

func unregisterCustomAction(name string) bool {
	id, ok := customActionNameToID[name]
	if !ok {
		return false
	}
	delete(customActionNameToID, name)
	delete(customActionCallbackAgents, id)
	return ok
}

func clearCustomAction() {
	customActionNameToID = make(map[string]uint64)
	customActionCallbackAgents = make(map[uint64]func(ctx *Context, taskId int64, actionName, customActionParam string, box Rect, recognitionDetail string) bool)
}

//export _MaaCustomActionCallbackAgent
func _MaaCustomActionCallbackAgent(
	ctx *C.MaaContext,
	taskId C.int64_t,
	actionName, customActionParam C.StringView,
	box C.ConstMaaRectPtr,
	recognitionDetail C.StringView,
	actionArg unsafe.Pointer,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(actionArg))
	callback := customActionCallbackAgents[id]
	curBoxRectBuffer := newRectBufferByHandle(unsafe.Pointer(box))

	ok := callback(
		&Context{handle: ctx},
		int64(taskId),
		C.GoString(actionName),
		C.GoString(customActionParam),
		curBoxRectBuffer.Get(),
		C.GoString(recognitionDetail),
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
