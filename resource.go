package maa

import (
	"sync"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/store"
)

type resourceStoreValue struct {
	NotificationCallbackID      uint64
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

var (
	resourceStore      = store.New[resourceStoreValue]()
	resourceStoreMutex sync.RWMutex
)

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
		uintptr(id),
	)
	if handle == 0 {
		return nil
	}

	resourceStoreMutex.Lock()
	resourceStore.Set(handle, resourceStoreValue{
		NotificationCallbackID:      id,
		CustomRecognizersCallbackID: make(map[string]uint64),
		CustomActionsCallbackID:     make(map[string]uint64),
	})
	resourceStoreMutex.Unlock()

	return &Resource{
		handle: handle,
	}
}

// Destroy frees the resource.
func (r *Resource) Destroy() {
	resourceStoreMutex.Lock()
	value := resourceStore.Get(r.handle)
	unregisterNotificationCallback(value.NotificationCallbackID)
	resourceStore.Del(r.handle)
	resourceStoreMutex.Unlock()

	maa.MaaResourceDestroy(r.handle)
}

func (r *Resource) setOption(key maa.MaaResOption, value unsafe.Pointer, valSize uintptr) bool {
	return maa.MaaResourceSetOption(
		r.handle,
		key,
		value,
		uint64(valSize),
	)
}

func (r *Resource) setInferenceDevice(device maa.MaaInferenceDevice) bool {
	return r.setOption(
		maa.MaaResOption_InferenceDevice,
		unsafe.Pointer(&device),
		unsafe.Sizeof(device),
	)
}

func (r *Resource) setInferenceExecutionProvider(ep maa.MaaInferenceExecutionProvider) bool {
	return r.setOption(
		maa.MaaResOption_InferenceExecutionProvider,
		unsafe.Pointer(&ep),
		unsafe.Sizeof(ep),
	)
}

func (r *Resource) setInference(ep maa.MaaInferenceExecutionProvider, deviceID maa.MaaInferenceDevice) bool {
	return r.setInferenceExecutionProvider(ep) && r.setInferenceDevice(deviceID)
}

// UseCPU
func (r *Resource) UseCPU() bool {
	return r.setInference(maa.MaaInferenceExecutionProvider_CPU, maa.MaaInferenceDevice_CPU)
}

type InterenceDevice = maa.MaaInferenceDevice

const (
	InterenceDeviceAuto int32 = -1
	InferenceDevice0    int32 = 0
	InferenceDevice1    int32 = 1
	// and more gpu id or flag...
)

// UseDirectml
func (r *Resource) UseDirectml(deviceID InterenceDevice) bool {
	return r.setInference(maa.MaaInferenceExecutionProvider_DirectML, deviceID)
}

// UseCoreml
func (r *Resource) UseCoreml(coremlFlag InterenceDevice) bool {
	return r.setInference(maa.MaaInferenceExecutionProvider_CoreML, coremlFlag)
}

// UseAutoExecutionProvider
func (r *Resource) UseAutoExecutionProvider() bool {
	return r.setInference(maa.MaaInferenceExecutionProvider_Auto, maa.MaaInferenceDevice_Auto)
}

// RegisterCustomRecognition registers a custom recognition to the resource.
func (r *Resource) RegisterCustomRecognition(name string, recognition CustomRecognition) bool {
	id := registerCustomRecognition(recognition)

	resourceStoreMutex.Lock()
	value := resourceStore.Get(r.handle)
	if oldID, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(oldID)
	}
	value.CustomRecognizersCallbackID[name] = id
	resourceStore.Set(r.handle, value)
	resourceStoreMutex.Unlock()

	return maa.MaaResourceRegisterCustomRecognition(
		r.handle,
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
}

// UnregisterCustomRecognition unregisters a custom recognition from the resource.
func (r *Resource) UnregisterCustomRecognition(name string) bool {
	resourceStoreMutex.Lock()
	defer resourceStoreMutex.Unlock()

	value := resourceStore.Get(r.handle)
	if id, ok := value.CustomRecognizersCallbackID[name]; ok {
		unregisterCustomRecognition(id)
	} else {
		return false
	}

	return maa.MaaResourceUnregisterCustomRecognition(r.handle, name)
}

// ClearCustomRecognition clears all custom recognitions registered from the resource.
func (r *Resource) ClearCustomRecognition() bool {
	resourceStoreMutex.Lock()
	defer resourceStoreMutex.Unlock()

	value := resourceStore.Get(r.handle)
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}

	return maa.MaaResourceClearCustomRecognition(r.handle)
}

// RegisterCustomAction registers a custom action to the resource.
func (r *Resource) RegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(action)

	resourceStoreMutex.Lock()
	value := resourceStore.Get(r.handle)
	if oldID, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(oldID)
	}
	value.CustomActionsCallbackID[name] = id
	resourceStore.Set(r.handle, value)
	resourceStoreMutex.Unlock()

	return maa.MaaResourceRegisterCustomAction(
		r.handle,
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
}

// UnregisterCustomAction unregisters a custom action from the resource.
func (r *Resource) UnregisterCustomAction(name string) bool {
	resourceStoreMutex.Lock()
	defer resourceStoreMutex.Unlock()

	value := resourceStore.Get(r.handle)
	if id, ok := value.CustomActionsCallbackID[name]; ok {
		unregisterCustomAction(id)
	} else {
		return false
	}

	return maa.MaaResourceUnregisterCustomAction(r.handle, name)
}

// ClearCustomAction clears all custom actions registered from the resource.
func (r *Resource) ClearCustomAction() bool {
	resourceStoreMutex.Lock()
	defer resourceStoreMutex.Unlock()

	value := resourceStore.Get(r.handle)
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}

	return maa.MaaResourceClearCustomAction(r.handle)
}

// PostBundle adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostBundle(path string) *Job {
	id := maa.MaaResourcePostBundle(r.handle, path)
	return NewJob(id, r.status, r.wait)
}

func (r *Resource) overridePipeline(override string) bool {
	return maa.MaaResourceOverridePipeline(r.handle, override)
}

// OverridePipeline overrides pipeline.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (r *Resource) OverridePipeline(override any) bool {
	if str, ok := override.(string); ok {
		return r.overridePipeline(str)
	}
	str, err := toJSON(override)
	if err != nil {
		return false
	}
	return r.overridePipeline(str)
}

// OverrideNext overrides the next list of task by name.
func (r *Resource) OverrideNext(name string, nextList []string) bool {
	list := buffer.NewStringListBuffer()
	defer list.Destroy()
	size := len(nextList)
	items := make([]*buffer.StringBuffer, size)
	for i := 0; i < size; i++ {
		items[i] = buffer.NewStringBuffer()
		items[i].Set(nextList[i])
		list.Append(items[i])
	}
	defer func() {
		for _, item := range items {
			item.Destroy()
		}
	}()
	return maa.MaaContextOverrideNext(r.handle, name, list.Handle())
}

// GetNodeJSON gets the node JSON by name.
func (r *Resource) GetNodeJSON(name string) (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := maa.MaaResourceGetNodeData(r.handle, name, buf.Handle())
	return buf.Get(), ok
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

	got := maa.MaaResourceGetHash(r.handle, hash.Handle())
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetNodeList returns the node list of the resource.
func (r *Resource) GetNodeList() ([]string, bool) {
	taskList := buffer.NewStringListBuffer()
	defer taskList.Destroy()

	got := maa.MaaResourceGetNodeList(r.handle, taskList.Handle())
	if !got {
		return []string{}, false
	}
	taskListArr := taskList.GetAll()

	return taskListArr, true
}

// AddSink adds a notification callback sink.
func (r *Resource) AddSink(notify Notification) bool {
	id := registerNotificationCallback(notify)
	return maa.MaaResourceAddSink(
		r.handle,
		_MaaNotificationCallbackAgent,
		uintptr(id),
	)
}

// RemoveSink removes a notification callback sink.
func (r *Resource) RemoveSink(notify Notification) bool {
	id := registerNotificationCallback(notify)
	defer unregisterNotificationCallback(id)
	return maa.MaaResourceRemoveSink(
		r.handle,
		_MaaNotificationCallbackAgent,
	)
}

// ClearSinks clears all notification callback sinks.
func (r *Resource) ClearSinks() bool {
	return maa.MaaResourceClearSinks(r.handle)
}
