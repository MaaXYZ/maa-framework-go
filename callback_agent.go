package maa

/*
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import "unsafe"

type callbackAgent struct {
	callback func(msg, detailsJson string)
}

//export _MaaAPICallbackAgent
func _MaaAPICallbackAgent(msg, detailsJson C.MaaStringView, callbackArg C.MaaTransparentArg) {
	agent := *(*callbackAgent)(unsafe.Pointer(callbackArg))
	if agent.callback == nil {
		return
	}
	agent.callback(C.GoString(msg), C.GoString(detailsJson))
}
