package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"unsafe"
)

type Resource struct {
	handle C.MaaResourceHandle
}

// NewResource creates a new resource.
//
// This function takes two arguments:
//
//   - callback: The callback function.
func NewResource(callback func(msg, detailsJson string)) *Resource {
	agent := &callbackAgent{callback: callback}
	handle := C.MaaResourceCreate(C.MaaAPICallback(C._MaaAPICallbackAgent), C.MaaTransparentArg(unsafe.Pointer(agent)))
	return &Resource{
		handle: handle,
	}
}

// Destroy frees the resource.
func (r *Resource) Destroy() {
	C.MaaResourceDestroy(r.handle)
}

// PostPath adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostPath(path string) int64 {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	return int64(C.MaaResourcePostPath(r.handle, cPath))
}

// Clear clears the resource loading paths.
// If the call is successful, it returns true. Otherwise, it returns false.
func (r *Resource) Clear() bool {
	return C.MaaResourceClear(r.handle) != 0
}

// Status gets the loading status of a resource identified by id.
func (r *Resource) Status(resId int64) Status {
	return Status(C.MaaResourceStatus(r.handle, C.int64_t(resId)))
}

// Wait waits for a resource to be loaded.
func (r *Resource) Wait(resId int64) Status {
	return Status(C.MaaResourceWait(r.handle, C.int64_t(resId)))
}

// Loaded checks if resources are loaded.
func (r *Resource) Loaded() bool {
	return C.MaaResourceLoaded(r.handle) != 0
}

// GetHash gets the hash of the resource.
func (r *Resource) GetHash() (string, bool) {
	hash := NewString()
	defer hash.Destroy()

	got := C.MaaResourceGetHash(r.handle, C.MaaStringBufferHandle(hash.Handle())) != 0
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetTaskList gets the task list of the resource.
func (r *Resource) GetTaskList() (string, bool) {
	taskList := NewString()
	defer taskList.Destroy()

	got := C.MaaResourceGetTaskList(r.handle, C.MaaStringBufferHandle(taskList.Handle())) != 0
	if !got {
		return "", false
	}
	return taskList.Get(), true
}
