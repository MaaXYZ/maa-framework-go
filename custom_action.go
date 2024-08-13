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
	"unsafe"
)

// CustomActionImpl defines an interface for custom action.
// Implementers of this interface must embed an CustomActionHandler struct
// and provide implementations for the Run and Stop methods.
type CustomActionImpl interface {
	Run(ctx SyncContext, taskName, ActionParam string, curBox buffer.Rect, curRecDetail string) bool
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
	act := *(*CustomActionImpl)(unsafe.Pointer(actionArg))
	curBoxRectBuffer := buffer.NewRectBufferByHandle(unsafe.Pointer(curBox))
	ok := act.Run(
		SyncContext{handle: ctx},
		C.GoString(taskName),
		C.GoString(customActionParam),
		curBoxRectBuffer.Get(),
		C.GoString(curRecDetail),
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAgent
func _StopAgent(actionArg C.MaaTransparentArg) {
	act := *(*CustomActionImpl)(unsafe.Pointer(actionArg))
	act.Stop()
}
