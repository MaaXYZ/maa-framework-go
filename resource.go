package maa

import (
	"encoding/json"
	"image"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/store"
)

type Resource struct {
	handle uintptr
}

// NewResource creates a new resource.
func NewResource() *Resource {
	handle := native.MaaResourceCreate()
	if handle == 0 {
		return nil
	}

	store.ResStore.Lock()
	store.ResStore.Set(handle, store.ResStoreValue{
		SinkIDToEventCallbackID:     make(map[int64]uint64),
		CustomRecognizersCallbackID: make(map[string]uint64),
		CustomActionsCallbackID:     make(map[string]uint64),
	})
	store.ResStore.Unlock()

	return &Resource{
		handle: handle,
	}
}

// Destroy frees the resource.
func (r *Resource) Destroy() {
	store.ResStore.Lock()
	value := store.ResStore.Get(r.handle)
	for _, id := range value.SinkIDToEventCallbackID {
		unregisterEventCallback(id)
	}
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}
	store.ResStore.Del(r.handle)
	store.ResStore.Unlock()

	native.MaaResourceDestroy(r.handle)
}

func (r *Resource) setOption(key native.MaaResOption, value unsafe.Pointer, valSize uintptr) bool {
	return native.MaaResourceSetOption(
		r.handle,
		key,
		value,
		uint64(valSize),
	)
}

func (r *Resource) setInferenceDevice(device native.MaaInferenceDevice) bool {
	return r.setOption(
		native.MaaResOption_InferenceDevice,
		unsafe.Pointer(&device),
		unsafe.Sizeof(device),
	)
}

func (r *Resource) setInferenceExecutionProvider(ep native.MaaInferenceExecutionProvider) bool {
	return r.setOption(
		native.MaaResOption_InferenceExecutionProvider,
		unsafe.Pointer(&ep),
		unsafe.Sizeof(ep),
	)
}

func (r *Resource) setInference(ep native.MaaInferenceExecutionProvider, deviceID native.MaaInferenceDevice) bool {
	return r.setInferenceExecutionProvider(ep) && r.setInferenceDevice(deviceID)
}

// UseCPU
func (r *Resource) UseCPU() bool {
	return r.setInference(native.MaaInferenceExecutionProvider_CPU, native.MaaInferenceDevice_CPU)
}

type InterenceDevice = native.MaaInferenceDevice

const (
	InterenceDeviceAuto int32 = -1
	InferenceDevice0    int32 = 0
	InferenceDevice1    int32 = 1
	// and more gpu id or flag...
)

// UseDirectml
func (r *Resource) UseDirectml(deviceID InterenceDevice) bool {
	return r.setInference(native.MaaInferenceExecutionProvider_DirectML, deviceID)
}

// UseCoreml
func (r *Resource) UseCoreml(coremlFlag InterenceDevice) bool {
	return r.setInference(native.MaaInferenceExecutionProvider_CoreML, coremlFlag)
}

// UseAutoExecutionProvider
func (r *Resource) UseAutoExecutionProvider() bool {
	return r.setInference(native.MaaInferenceExecutionProvider_Auto, native.MaaInferenceDevice_Auto)
}

// RegisterCustomRecognition registers a custom recognition runner to the resource.
func (r *Resource) RegisterCustomRecognition(name string, recognition CustomRecognitionRunner) bool {
	id := registerCustomRecognition(recognition)

	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		if oldID, ok := v.CustomRecognizersCallbackID[name]; ok {
			unregisterCustomRecognition(oldID)
		}
		v.CustomRecognizersCallbackID[name] = id
	})

	return native.MaaResourceRegisterCustomRecognition(
		r.handle,
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
}

// UnregisterCustomRecognition unregisters a custom recognition runner from the resource.
func (r *Resource) UnregisterCustomRecognition(name string) bool {
	var found bool
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		if id, ok := v.CustomRecognizersCallbackID[name]; ok {
			unregisterCustomRecognition(id)
			delete(v.CustomRecognizersCallbackID, name)
			found = true
		}
	})
	if !found {
		return false
	}
	return native.MaaResourceUnregisterCustomRecognition(r.handle, name)
}

// ClearCustomRecognition clears all custom recognitions runner registered from the resource.
func (r *Resource) ClearCustomRecognition() bool {
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		for _, id := range v.CustomRecognizersCallbackID {
			unregisterCustomRecognition(id)
		}
		v.CustomRecognizersCallbackID = make(map[string]uint64)
	})

	return native.MaaResourceClearCustomRecognition(r.handle)
}

// RegisterCustomAction registers a custom action runner to the resource.
func (r *Resource) RegisterCustomAction(name string, action CustomActionRunner) bool {
	id := registerCustomAction(action)

	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		if oldID, ok := v.CustomActionsCallbackID[name]; ok {
			unregisterCustomAction(oldID)
		}
		v.CustomActionsCallbackID[name] = id
	})

	return native.MaaResourceRegisterCustomAction(
		r.handle,
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
}

// UnregisterCustomAction unregisters a custom action runner from the resource.
func (r *Resource) UnregisterCustomAction(name string) bool {
	var found bool
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		if id, ok := v.CustomActionsCallbackID[name]; ok {
			unregisterCustomAction(id)
			delete(v.CustomActionsCallbackID, name)
			found = true
		}
	})
	if !found {
		return false
	}
	return native.MaaResourceUnregisterCustomAction(r.handle, name)
}

// ClearCustomAction clears all custom actions runners registered from the resource.
func (r *Resource) ClearCustomAction() bool {
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		for _, id := range v.CustomActionsCallbackID {
			unregisterCustomAction(id)
		}
		v.CustomActionsCallbackID = make(map[string]uint64)
	})

	return native.MaaResourceClearCustomAction(r.handle)
}

// PostBundle adds a path to the resource loading paths.
// Return id of the resource.
func (r *Resource) PostBundle(path string) *Job {
	id := native.MaaResourcePostBundle(r.handle, path)
	return newJob(id, r.status, r.wait)
}

// PostOcrModel adds an OCR model to the resource loading paths.
func (r *Resource) PostOcrModel(path string) *Job {
	id := native.MaaResourcePostOcrModel(r.handle, path)
	return newJob(id, r.status, r.wait)
}

// PostPipeline adds a pipeline to the resource loading paths.
func (r *Resource) PostPipeline(path string) *Job {
	id := native.MaaResourcePostPipeline(r.handle, path)
	return newJob(id, r.status, r.wait)
}

// PostImage adds an image to the resource loading paths.
func (r *Resource) PostImage(path string) *Job {
	id := native.MaaResourcePostImage(r.handle, path)
	return newJob(id, r.status, r.wait)
}

func (r *Resource) overridePipeline(override string) bool {
	return native.MaaResourceOverridePipeline(r.handle, override)
}

// OverridePipeline overrides pipeline.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (r *Resource) OverridePipeline(override any) bool {
	switch v := override.(type) {
	case string:
		return r.overridePipeline(v)
	case []byte:
		return r.overridePipeline(string(v))
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return false
		}
		return r.overridePipeline(string(jsonBytes))
	}
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
	return native.MaaContextOverrideNext(r.handle, name, list.Handle())
}

func (r *Resource) OverriderImage(imageName string, image image.Image) bool {
	img := buffer.NewImageBuffer()
	defer img.Destroy()
	img.Set(image)
	return native.MaaResourceOverrideImage(r.handle, imageName, img.Handle())
}

// GetNodeJSON gets the node JSON by name.
func (r *Resource) GetNodeJSON(name string) (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaResourceGetNodeData(r.handle, name, buf.Handle())
	return buf.Get(), ok
}

// Clear clears the resource loading paths.
func (r *Resource) Clear() bool {
	return native.MaaResourceClear(r.handle)
}

// status returns the loading status of a resource identified by id.
func (r *Resource) status(resId int64) Status {
	return Status(native.MaaResourceStatus(r.handle, resId))
}

func (r *Resource) wait(resId int64) Status {
	return Status(native.MaaResourceWait(r.handle, resId))
}

// Loaded checks if resources are loaded.
func (r *Resource) Loaded() bool {
	return native.MaaResourceLoaded(r.handle)
}

// GetHash returns the hash of the resource.
func (r *Resource) GetHash() (string, bool) {
	hash := buffer.NewStringBuffer()
	defer hash.Destroy()

	got := native.MaaResourceGetHash(r.handle, hash.Handle())
	if !got {
		return "", false
	}
	return hash.Get(), true
}

// GetNodeList returns the node list of the resource.
func (r *Resource) GetNodeList() ([]string, bool) {
	taskList := buffer.NewStringListBuffer()
	defer taskList.Destroy()

	got := native.MaaResourceGetNodeList(r.handle, taskList.Handle())
	if !got {
		return []string{}, false
	}
	taskListArr := taskList.GetAll()

	return taskListArr, true
}

// GetCustomRecognitionList returns the custom recognition list of the resource.
func (r *Resource) GetCustomRecognitionList() ([]string, bool) {
	recognitionList := buffer.NewStringListBuffer()
	defer recognitionList.Destroy()

	got := native.MaaResourceGetCustomRecognitionList(r.handle, recognitionList.Handle())
	if !got {
		return []string{}, false
	}

	return recognitionList.GetAll(), true
}

// GetCustomActionList returns the custom action list of the resource.
func (r *Resource) GetCustomActionList() ([]string, bool) {
	actionList := buffer.NewStringListBuffer()
	defer actionList.Destroy()

	got := native.MaaResourceGetCustomActionList(r.handle, actionList.Handle())
	if !got {
		return []string{}, false
	}

	return actionList.GetAll(), true
}

// AddSink adds a event callback sink and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (r *Resource) AddSink(sink ResourceEventSink) int64 {
	id := registerEventCallback(sink)
	sinkId := native.MaaResourceAddSink(
		r.handle,
		_MaaEventCallbackAgent,
		uintptr(id),
	)

	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		v.SinkIDToEventCallbackID[sinkId] = id
	})

	return sinkId
}

// RemoveSink removes a event callback sink by sink ID.
func (r *Resource) RemoveSink(sinkId int64) {
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		unregisterEventCallback(v.SinkIDToEventCallbackID[sinkId])
		delete(v.SinkIDToEventCallbackID, sinkId)
	})

	native.MaaResourceRemoveSink(r.handle, sinkId)
}

// ClearSinks clears all event callback sinks.
func (r *Resource) ClearSinks() {
	store.ResStore.Update(r.handle, func(v *store.ResStoreValue) {
		for _, id := range v.SinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.SinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaResourceClearSinks(r.handle)
}

type ResourceEventSink interface {
	OnResourceLoading(res *Resource, event EventStatus, detail ResourceLoadingDetail)
}

// ResourceEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type ResourceEventSinkAdapter struct {
	onResourceLoading func(EventStatus, ResourceLoadingDetail)
}

func (a *ResourceEventSinkAdapter) OnResourceLoading(res *Resource, status EventStatus, detail ResourceLoadingDetail) {
	if a == nil || a.onResourceLoading == nil {
		return
	}
	a.onResourceLoading(status, detail)
}

// OnResourceLoading registers a callback sink that only handles Resource.Loading events and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (r *Resource) OnResourceLoading(fn func(EventStatus, ResourceLoadingDetail)) int64 {
	sink := &ResourceEventSinkAdapter{onResourceLoading: fn}
	return r.AddSink(sink)
}
