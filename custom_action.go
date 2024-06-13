package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_action.h"

extern uint8_t _RunAgent(
	MaaSyncContextHandle SyncCtx,
	_GoString_ taskName,
	_GoString_ customActionParam,
	MaaRectHandle curBox ,
	_GoString_ curRecDetail,
	MaaTransparentArg actionArg);

extern void _StopAgent(MaaTransparentArg actionArg);
*/
import "C"
import (
	"unsafe"
)

// CustomAction defines an interface for custom action.
// Implementers of this interface must embed an Action struct
// and provide implementations for the Run and Stop methods.
type CustomAction interface {
	Run(ctx SyncContext, taskName, ActionParam string, curBox RectBuffer, curRecDetail string) bool
	Stop()

	Handle() unsafe.Pointer
	Destroy()
}

type Action struct {
	handle C.MaaCustomActionHandle
}

func NewAction() Action {
	return Action{handle: C.MaaCustomActionHandleCreate(C.RunCallback(C._RunAgent), C.StopCallback(C._StopAgent))}
}

func (a Action) Handle() unsafe.Pointer {
	return unsafe.Pointer(a.handle)
}

func (a Action) Destroy() {
	C.MaaCustomActionHandleDestroy(a.handle)
}

//export _RunAgent
func _RunAgent(
	SyncCtx C.MaaSyncContextHandle,
	taskName, customActionParam string,
	curBox C.MaaRectHandle,
	curRecDetail string,
	actionArg unsafe.Pointer,
) C.uint8_t {
	if actionArg == nil {
		return C.uint8_t(0)
	}

	act := *(*CustomAction)(actionArg)
	ok := act.Run(
		SyncContext(SyncCtx),
		taskName,
		customActionParam,
		&rectBuffer{handle: curBox},
		curRecDetail,
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAgent
func _StopAgent(actionArg unsafe.Pointer) {
	if actionArg == nil {
		return
	}

	act := *(*CustomAction)(actionArg)
	act.Stop()
}
