package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_action.h"

extern uint8_t _RunAgent(
	MaaSyncContextHandle SyncCtx,
	MaaStringView taskName,
	MaaStringView customActionParam,
	MaaRectHandle curBox ,
	MaaStringView curRecDetail,
	MaaTransparentArg actionArg);

extern void _StopAgent(MaaTransparentArg actionArg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/buffer"
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

func unregisterCustomAction(name string) bool {
	id, ok := customActionNameToID[name]
	if !ok {
		return false
	}
	delete(customActionNameToID, name)
	delete(customActionAgents, id)
	return ok
}

func clearCustomAction() {
	customActionNameToID = make(map[string]uint64)
	customActionAgents = make(map[uint64]CustomAction)
}

// CustomAction defines an interface for custom action.
// Implementers of this interface must embed an CustomActionHandler struct
// and provide implementations for the Run and Stop methods.
type CustomAction interface {
	Run(ctx SyncContext, taskName, ActionParam string, curBox Rect, curRecDetail string) bool
	Stop()

	Handle() unsafe.Pointer
	Destroy()
}

type CustomActionHandler struct {
	handle C.MaaCustomActionHandle
}

func NewCustomActionHandler() CustomActionHandler {
	return CustomActionHandler{handle: C.MaaCustomActionHandleCreate(C.RunCallback(C._RunAgent), C.StopCallback(C._StopAgent))}
}

func (a CustomActionHandler) Handle() unsafe.Pointer {
	return unsafe.Pointer(a.handle)
}

func (a CustomActionHandler) Destroy() {
	C.MaaCustomActionHandleDestroy(a.handle)
}

//export _RunAgent
func _RunAgent(
	ctx C.MaaSyncContextHandle,
	taskName, customActionParam C.MaaStringView,
	curBox C.MaaRectHandle,
	curRecDetail C.MaaStringView,
	actionArg C.MaaTransparentArg,
) C.uint8_t {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(unsafe.Pointer(actionArg)))
	act := customActionAgents[id]
	curBoxRectBuffer := buffer.NewRectBufferByHandle(unsafe.Pointer(curBox))

	ok := act.Run(
		SyncContext{handle: ctx},
		C.GoString(taskName),
		C.GoString(customActionParam),
		toMaaRect(curBoxRectBuffer.Get()),
		C.GoString(curRecDetail),
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAgent
func _StopAgent(actionArg C.MaaTransparentArg) {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(unsafe.Pointer(actionArg)))
	act := customActionAgents[id]
	act.Stop()
}
