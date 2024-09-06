package maa

/*
#include <MaaFramework/MaaAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

var (
	notificationCallbackID     uint64
	notificationCallbackAgents = make(map[uint64]func(msg, detailsJson string))
)

func registerNotificationCallback(callback func(msg, detailsJson string)) uint64 {
	id := atomic.AddUint64(&notificationCallbackID, 1)
	notificationCallbackAgents[id] = callback
	return id
}

//export _MaaNotificationCallbackAgent
func _MaaNotificationCallbackAgent(msg, detailsJson C.CString, callbackArg unsafe.Pointer) {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(uintptr(callbackArg))
	callback := notificationCallbackAgents[id]
	if callback == nil {
		return
	}
	callback(C.GoString(msg), C.GoString(detailsJson))
}
