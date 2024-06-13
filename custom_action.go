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

// CustomActionImpl defines an interface for custom action.
// Implementers of this interface must embed an CustomActionHandler struct
// and provide implementations for the Run and Stop methods.
type CustomActionImpl interface {
	Run(ctx SyncContext, taskName, ActionParam string, curBox RectBuffer, curRecDetail string) bool
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
	SyncCtx C.MaaSyncContextHandle,
	taskName, customActionParam string,
	curBox C.MaaRectHandle,
	curRecDetail string,
	actionArg unsafe.Pointer,
) C.uint8_t {
	if actionArg == nil {
		return C.uint8_t(0)
	}

	act := *(*CustomActionImpl)(actionArg)
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

	act := *(*CustomActionImpl)(actionArg)
	act.Stop()
}
