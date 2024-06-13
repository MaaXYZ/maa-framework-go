package maa

/*
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(_GoString_ msg, _GoString_ detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"unsafe"
)

type callbackAgent struct {
	callback func(msg, detailsJson string)
}

//export _MaaAPICallbackAgent
func _MaaAPICallbackAgent(msg, detailsJson string, callbackArg unsafe.Pointer) {
	agent := (*callbackAgent)(callbackArg)
	if agent.callback == nil {
		return
	}
	agent.callback(msg, detailsJson)
}
