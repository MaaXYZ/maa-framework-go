package maa

/*
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

var (
	callbackID     uint64
	callbackAgents = make(map[uint64]func(msg, detailsJson string))
)

func registerCallback(callback func(msg, detailsJson string)) uint64 {
	id := atomic.AddUint64(&callbackID, 1)
	callbackAgents[id] = callback
	return id
}

//export _MaaAPICallbackAgent
func _MaaAPICallbackAgent(msg, detailsJson C.MaaStringView, callbackArg C.MaaTransparentArg) {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(unsafe.Pointer(callbackArg)))
	callback := callbackAgents[id]
	if callback == nil {
		return
	}
	callback(C.GoString(msg), C.GoString(detailsJson))
}
