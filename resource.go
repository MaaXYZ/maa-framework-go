package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);

extern uint8_t _MaaCustomRecognitionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_recognizer_name,
	const char* custom_recognition_param,
	const MaaImageBuffer* image,
	const MaaRect* roi,
	void* recognizer_arg,
	MaaRect* out_box,
	MaaStringBuffer* out_detail);

extern uint8_t _MaaCustomActionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_action_name,
	const char* custom_action_param,
	int64_t rec_id,
	const MaaRect* box ,
	void* actionArg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/store"
	"unsafe"
)

type resourceStoreValue struct {
	NotificationCallbackID      uint64
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

var resourceStore = store.New[resourceStoreValue]()

type Resource struct {
	handle *C.MaaResource
}

// NewResource creates a new resource.
func NewResource(notify Notification) *Resource {
	id := registerNotificationCallback(notify)
	handle := C.MaaResourceCreate(
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	resourceStore.Set(unsafe.Pointer(handle), resourceStoreValue{
		NotificationCallbackID:      id,
		CustomRecognizersCallbackID: make(map[string]uint64),
		CustomActionsCallbackID:     make(map[string]uint64),
	})
	return &Resource{
		handle: handle,
	}
}

// Destroy frees the resource.
func (r *Resource) Destroy() {
	value := resourceStore.Get(r.Handle())
	unregisterNotificationCallback(value.NotificationCallbackID)
	resourceStore.Del(r.Handle())
	C.MaaResourceDestroy(r.handle)
}

func (r *Resource) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

// RegisterCustomRecognition registers a custom recognition to the resource.
func (r *Resource) RegisterCustomRecognition(name string, recognition CustomRecognition) bool {
	id := registerCustomRecognition(recognition)
	value := resourceStore.Get(r.Handle())
	if oldID, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(oldID)
	}
	value.CustomRecognizersCallbackID[name] = id
	resourceStore.Set(r.Handle(), value)

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceRegisterCustomRecognition(
		r.handle,
		cName,
		C.MaaCustomRecognitionCallback(C._MaaCustomRecognitionCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got != 0
}

// UnregisterCustomRecognition unregisters a custom recognition from the resource.
func (r *Resource) UnregisterCustomRecognition(name string) bool {
	value := resourceStore.Get(r.Handle())
	if id, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(id)
	} else {
		return false
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceUnregisterCustomRecognition(r.handle, cName)
	return got != 0
}

// ClearCustomRecognition clears all custom recognitions registered from the resource.
func (r *Resource) ClearCustomRecognition() bool {
	value := resourceStore.Get(r.Handle())
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}

	got := C.MaaResourceClearCustomRecognition(r.handle)
	return got != 0
}

// RegisterCustomAction registers a custom action to the resource.
func (r *Resource) RegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(action)
	value := resourceStore.Get(r.Handle())
	if oldID, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(oldID)
	}
	value.CustomActionsCallbackID[name] = id
	resourceStore.Set(r.Handle(), value)

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
	value := resourceStore.Get(r.Handle())
	if id, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(id)
	} else {
		return false
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	got := C.MaaResourceUnregisterCustomAction(r.handle, cName)
	return got != 0
}

// ClearCustomAction clears all custom actions registered from the resource.
func (r *Resource) ClearCustomAction() bool {
	value := resourceStore.Get(r.Handle())
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}

	got := C.MaaResourceClearCustomAction(r.handle)
	return got != 0
}

// PostPath adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostPath(path string) *Job {
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
	hash := newStringBuffer()
	defer hash.Destroy()

	got := C.MaaResourceGetHash(r.handle, (*C.MaaStringBuffer)(hash.Handle())) != 0
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetTaskList returns the task list of the resource.
func (r *Resource) GetTaskList() ([]string, bool) {
	taskList := newStringListBuffer()
	defer taskList.Destroy()

	got := C.MaaResourceGetTaskList(r.handle, (*C.MaaStringListBuffer)(taskList.Handle())) != 0
	if !got {
		return []string{}, false
	}
	taskListArr := taskList.GetAll()

	return taskListArr, true
}
