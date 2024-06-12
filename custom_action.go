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

type CustomAction struct {
	handle C.MaaCustomActionHandle

	run func(ctx SyncContext,
		taskName, customActionParam string,
		curBox RectBuffer,
		curRecDetail string,
		actionArg interface{},
	) bool
	stop func(actionArg interface{})
}

func (a *CustomAction) Set(
	run func(ctx SyncContext,
		taskName, customActionParam string,
		curBox RectBuffer,
		curRecDetail string,
		actionArg interface{},
	) bool,
	stop func(actionArg interface{}),
) {
	a.run = run
	a.stop = stop
	a.handle = C.MaaCustomActionHandleCreate(C.RunCallback(C._RunAgent), C.StopCallback(C._StopAgent))
}

func (a *CustomAction) Handle() unsafe.Pointer {
	return unsafe.Pointer(a.handle)
}

func (a *CustomAction) Destroy() {
	C.MaaCustomActionHandleDestroy(a.handle)
}

type customActionAgent struct {
	act *CustomAction
	arg interface{}
}

//export _RunAgent
func _RunAgent(
	SyncCtx C.MaaSyncContextHandle,
	taskName, customActionParam string,
	curBox C.MaaRectHandle,
	curRecDetail string,
	actionArg unsafe.Pointer,
) C.uint8_t {
	agent := (*customActionAgent)(actionArg)
	act := agent.act
	ok := act.run(
		SyncContext(SyncCtx),
		taskName,
		customActionParam,
		&rectBuffer{handle: curBox},
		curRecDetail,
		agent.arg,
	)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAgent
func _StopAgent(actionArg unsafe.Pointer) {
	agent := (*customActionAgent)(actionArg)
	act := agent.act
	act.stop(agent.arg)
}
