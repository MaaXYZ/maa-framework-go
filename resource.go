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
func (r *Resource) PostPath(path string) Job {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	id := int64(C.MaaResourcePostPath(r.handle, cPath))
	return NewJob(id, r.status)
}

// Clear clears the resource loading paths.
func (r *Resource) Clear() bool {
	return C.MaaResourceClear(r.handle) != 0
}

// status returns the loading status of a resource identified by id.
func (r *Resource) status(resId int64) Status {
	return Status(C.MaaResourceStatus(r.handle, C.int64_t(resId)))
}

// Loaded checks if resources are loaded.
func (r *Resource) Loaded() bool {
	return C.MaaResourceLoaded(r.handle) != 0
}

// GetHash returns the hash of the resource.
func (r *Resource) GetHash() (string, bool) {
	hash := NewStringBuffer()
	defer hash.Destroy()

	got := C.MaaResourceGetHash(r.handle, C.MaaStringBufferHandle(hash.Handle())) != 0
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetTaskList returns the task list of the resource.
func (r *Resource) GetTaskList() (string, bool) {
	taskList := NewStringBuffer()
	defer taskList.Destroy()

	got := C.MaaResourceGetTaskList(r.handle, C.MaaStringBufferHandle(taskList.Handle())) != 0
	if !got {
		return "", false
	}
	return taskList.Get(), true
}
