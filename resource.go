package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);

extern uint8_t _MaaCustomRecognizerCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* recognizer_name,
	const char* custom_recognition_param,
	const MaaImageBuffer* image,
	void* recognizer_arg,
	MaaRect* out_box,
	MaaStringBuffer* out_detail);

extern uint8_t _MaaCustomActionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char*  task_name,
	const char*  customActionParam,
	MaaRect* box ,
	const char* recognition_detail,
	void* actionArg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"unsafe"
)

type Resource struct {
	handle *C.MaaResource
}

// NewResource creates a new resource.
func NewResource(callback func(msg, detailsJson string)) *Resource {
	id := registerNotificationCallback(callback)
	handle := C.MaaResourceCreate(
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	return &Resource{
		handle: handle,
	}
}

// Destroy frees the resource.
func (r *Resource) Destroy() {
	C.MaaResourceDestroy(r.handle)
}

// RegisterCustomRecognizer registers a custom recognizer to the resource.
func (r *Resource) RegisterCustomRecognizer(name string, recognizer CustomRecognizer) bool {
	id := registerCustomRecognizer(name, recognizer)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceRegisterCustomRecognizer(
		r.handle,
		cName,
		C.MaaCustomRecognizerCallback(C._MaaCustomRecognizerCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got != 0
}

// UnregisterCustomRecognizer unregisters a custom recognizer from the resource.
func (r *Resource) UnregisterCustomRecognizer(name string) bool {
	unregisterCustomRecognizer(name)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceUnregisterCustomRecognizer(r.handle, cName)
	return got != 0
}

// ClearCustomRecognizer clears all custom recognizers registered from the resource.
func (r *Resource) ClearCustomRecognizer() bool {
	clearCustomRecognizer()

	got := C.MaaResourceClearCustomRecognizer(r.handle)
	return got != 0
}

// RegisterCustomAction registers a custom action to the resource.
func (r *Resource) RegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(name, action)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceRegisterCustomAction(
		r.handle,
		cName,
		C.MaaCustomActionCallback(C._MaaCustomActionCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got != 0
}

// UnregisterCustomAction unregisters a custom action from the resource.
func (r *Resource) UnregisterCustomAction(name string) bool {
	unregisterCustomAction(name)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceUnregisterCustomAction(r.handle, cName)
	return got != 0
}

// ClearCustomAction clears all custom actions registered from the resource.
func (r *Resource) ClearCustomAction() bool {
	clearCustomAction()

	got := C.MaaResourceClearCustomAction(r.handle)
	return got != 0
}

// PostPath adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostPath(path string) Job {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	id := int64(C.MaaResourcePostPath(r.handle, cPath))
	return NewJob(id, r.status, r.wait)
}

// Clear clears the resource loading paths.
func (r *Resource) Clear() bool {
	return C.MaaResourceClear(r.handle) != 0
}

// status returns the loading status of a resource identified by id.
func (r *Resource) status(resId int64) Status {
	return Status(C.MaaResourceStatus(r.handle, C.int64_t(resId)))
}

func (r *Resource) wait(resId int64) Status {
	return Status(C.MaaResourceWait(r.handle, C.int64_t(resId)))
}

// Loaded checks if resources are loaded.
func (r *Resource) Loaded() bool {
	return C.MaaResourceLoaded(r.handle) != 0
}

// GetHash returns the hash of the resource.
func (r *Resource) GetHash() (string, bool) {
	hash := buffer.NewStringBuffer()
	defer hash.Destroy()

	got := C.MaaResourceGetHash(r.handle, (*C.MaaStringBuffer)(hash.Handle())) != 0
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetTaskList returns the task list of the resource.
func (r *Resource) GetTaskList() ([]string, bool) {
	taskList := buffer.NewStringListBuffer()
	defer taskList.Destroy()

	got := C.MaaResourceGetTaskList(r.handle, (*C.MaaStringListBuffer)(taskList.Handle())) != 0
	if !got {
		return []string{}, false
	}
	taskListArr := taskList.GetAll()

	return taskListArr, true
}
