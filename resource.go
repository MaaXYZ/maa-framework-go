package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/internal/maa"
	"github.com/MaaXYZ/maa-framework-go/internal/store"
)

type resourceStoreValue struct {
	NotificationCallbackID      uint64
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

var resourceStore = store.New[resourceStoreValue]()

type Resource struct {
	handle uintptr
}

// NewResource creates a new resource.
func NewResource(notify Notification) *Resource {
	id := registerNotificationCallback(notify)
	handle := maa.MaaResourceCreate(
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == 0 {
		return nil
	}
	resourceStore.Set(handle, resourceStoreValue{
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
	value := resourceStore.Get(r.handle)
	unregisterNotificationCallback(value.NotificationCallbackID)
	resourceStore.Del(r.handle)
	maa.MaaResourceDestroy(r.handle)
}

func (r *Resource) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *Resource) setOption(key maa.MaaResOption, value unsafe.Pointer, valSize uintptr) bool {
	return maa.MaaResourceSetOption(
		r.handle,
		key,
		value,
		uint64(valSize),
	)
}

type InterfaceDevice int32

// InterfaceDevice
const (
	InterfaceDeviceCPU  InterfaceDevice = -2
	InterfaceDeviceAuto InterfaceDevice = -1
	InterfaceDeviceGPU0 InterfaceDevice = 0
	interfaceDeviceGPU1 InterfaceDevice = 1
	// and more gpu id...
)

func (r *Resource) SetInterfaceDevice(device InterfaceDevice) bool {
	return r.setOption(
		maa.MaaResOption_InterfaceDevice,
		unsafe.Pointer(&device),
		unsafe.Sizeof(device),
	)
}

// RegisterCustomRecognition registers a custom recognition to the resource.
func (r *Resource) RegisterCustomRecognition(name string, recognition CustomRecognition) bool {
	id := registerCustomRecognition(recognition)
	value := resourceStore.Get(r.handle)
	if oldID, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(oldID)
	}
	value.CustomRecognizersCallbackID[name] = id
	resourceStore.Set(r.handle, value)

	got := maa.MaaResourceRegisterCustomRecognition(
		r.handle,
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got
}

// UnregisterCustomRecognition unregisters a custom recognition from the resource.
func (r *Resource) UnregisterCustomRecognition(name string) bool {
	value := resourceStore.Get(r.handle)
	if id, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(id)
	} else {
		return false
	}

	got := maa.MaaResourceUnregisterCustomRecognition(r.handle, name)
	return got
}

// ClearCustomRecognition clears all custom recognitions registered from the resource.
func (r *Resource) ClearCustomRecognition() bool {
	value := resourceStore.Get(r.handle)
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}

	got := maa.MaaResourceClearCustomRecognition(r.handle)
	return got
}

// RegisterCustomAction registers a custom action to the resource.
func (r *Resource) RegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(action)
	value := resourceStore.Get(r.handle)
	if oldID, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(oldID)
	}
	value.CustomActionsCallbackID[name] = id
	resourceStore.Set(r.handle, value)

	got := maa.MaaResourceRegisterCustomAction(
		r.handle,
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got
}

// UnregisterCustomAction unregisters a custom action from the resource.
func (r *Resource) UnregisterCustomAction(name string) bool {
	value := resourceStore.Get(r.handle)
	if id, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(id)
	} else {
		return false
	}

	got := maa.MaaResourceUnregisterCustomAction(r.handle, name)
	return got
}

// ClearCustomAction clears all custom actions registered from the resource.
func (r *Resource) ClearCustomAction() bool {
	value := resourceStore.Get(r.handle)
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}

	got := maa.MaaResourceClearCustomAction(r.handle)
	return got
}

// PostPath adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostPath(path string) *Job {
	id := maa.MaaResourcePostPath(r.handle, path)
	return NewJob(id, r.status, r.wait)
}

// Clear clears the resource loading paths.
func (r *Resource) Clear() bool {
	return maa.MaaResourceClear(r.handle)
}

// status returns the loading status of a resource identified by id.
func (r *Resource) status(resId int64) Status {
	return Status(maa.MaaResourceStatus(r.handle, resId))
}

func (r *Resource) wait(resId int64) Status {
	return Status(maa.MaaResourceWait(r.handle, resId))
}

// Loaded checks if resources are loaded.
func (r *Resource) Loaded() bool {
	return maa.MaaResourceLoaded(r.handle)
}

// GetHash returns the hash of the resource.
func (r *Resource) GetHash() (string, bool) {
	hash := buffer.NewStringBuffer()
	defer hash.Destroy()

	got := maa.MaaResourceGetHash(r.handle, uintptr(hash.Handle()))
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetTaskList returns the task list of the resource.
func (r *Resource) GetTaskList() ([]string, bool) {
	taskList := buffer.NewStringListBuffer()
	defer taskList.Destroy()

	got := maa.MaaResourceGetTaskList(r.handle, uintptr(taskList.Handle()))
	if !got {
		return []string{}, false
	}
	taskListArr := taskList.GetAll()

	return taskListArr, true
}
