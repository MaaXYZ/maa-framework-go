package maa

import (
	"encoding/json"
	"image"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/store"
)

type Tasker struct {
	handle uintptr
}

// NewTasker creates a new tasker.
func NewTasker() *Tasker {
	handle := native.MaaTaskerCreate()
	if handle == 0 {
		return nil
	}

	store.TaskerStore.Lock()
	store.TaskerStore.Set(handle, store.TaskerStoreValue{
		SinkIDToEventCallbackID:        make(map[int64]uint64),
		ContextSinkIDToEventCallbackID: make(map[int64]uint64),
	})
	store.TaskerStore.Unlock()

	return &Tasker{handle: handle}
}

// Destroy free the tasker.
func (t *Tasker) Destroy() {
	store.TaskerStore.Lock()
	value := store.TaskerStore.Get(t.handle)
	for _, id := range value.SinkIDToEventCallbackID {
		unregisterEventCallback(id)
	}
	for _, id := range value.ContextSinkIDToEventCallbackID {
		unregisterEventCallback(id)
	}
	store.TaskerStore.Del(t.handle)
	store.TaskerStore.Unlock()

	native.MaaTaskerDestroy(t.handle)
}

// BindResource binds the tasker to an initialized resource.
func (t *Tasker) BindResource(res *Resource) bool {
	return native.MaaTaskerBindResource(t.handle, res.handle)
}

// BindController binds the tasker to an initialized controller.
func (t *Tasker) BindController(ctrl *Controller) bool {
	return native.MaaTaskerBindController(t.handle, ctrl.handle)
}

// Initialized checks if the tasker is initialized.
func (t *Tasker) Initialized() bool {
	return native.MaaTaskerInited(t.handle)
}

func (t *Tasker) handleOverride(entry string, postFunc func(entry, override string) *TaskJob, override ...any) *TaskJob {
	if len(override) == 0 {
		return postFunc(entry, "{}")
	}

	overrideValue := override[0]
	switch v := overrideValue.(type) {
	case string:
		return postFunc(entry, v)
	case []byte:
		return postFunc(entry, string(v))
	default:
		if v == nil {
			return postFunc(entry, "{}")
		}

		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return postFunc(entry, "{}")
		}
		return postFunc(entry, string(jsonBytes))
	}
}

func (t *Tasker) postTask(entry, pipelineOverride string) *TaskJob {
	id := native.MaaTaskerPostTask(t.handle, entry, pipelineOverride)
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// PostTask posts a task to the tasker.
// `override` is an optional parameter. If provided, it should be a single value
// that can be a JSON string or any data type that can be marshaled to JSON.
// If multiple values are provided, only the first one will be used.
func (t *Tasker) PostTask(entry string, override ...any) *TaskJob {
	return t.handleOverride(entry, t.postTask, override...)
}

// PostRecognition posts a recognition to the tasker.
func (t *Tasker) PostRecognition(recType NodeRecognitionType, recParam NodeRecognitionParam, img image.Image) *TaskJob {
	imgBuf := buffer.NewImageBuffer()
	defer imgBuf.Destroy()
	imgBuf.Set(img)

	recParamJSON, _ := json.Marshal(recParam)

	id := native.MaaTaskerPostRecognition(t.handle, string(recType), string(recParamJSON), imgBuf.Handle())
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// PostAction posts an action to the tasker.
func (t *Tasker) PostAction(actionType NodeActionType, actionParam NodeActionParam, box Rect, recoDetail *RecognitionDetail) *TaskJob {
	rectBuf := buffer.NewRectBuffer()
	defer rectBuf.Destroy()
	rectBuf.Set(box)

	actParamJSON, _ := json.Marshal(actionParam)
	recoDetailJSON, _ := json.Marshal(recoDetail)

	id := native.MaaTaskerPostAction(t.handle, string(actionType), string(actParamJSON), rectBuf.Handle(), string(recoDetailJSON))
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// Stopping checks whether the tasker is stopping.
func (t *Tasker) Stopping() bool {
	return native.MaaTaskerStopping(t.handle)
}

// status returns the status of a task identified by the id.
func (t *Tasker) status(id int64) Status {
	return Status(native.MaaTaskerStatus(t.handle, id))
}

// wait waits until the task is complete and returns the status of the completed task identified by the id.
func (t *Tasker) wait(id int64) Status {
	return Status(native.MaaTaskerWait(t.handle, id))
}

// Running checks if the instance running.
func (t *Tasker) Running() bool {
	return native.MaaTaskerRunning(t.handle)
}

// PostStop posts a stop signal to the tasker.
func (t *Tasker) PostStop() *TaskJob {
	id := native.MaaTaskerPostStop(t.handle)
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// GetResource returns the resource handle of the tasker.
func (t *Tasker) GetResource() *Resource {
	handle := native.MaaTaskerGetResource(t.handle)
	return &Resource{handle: handle}
}

// GetController returns the controller handle of the tasker.
func (t *Tasker) GetController() *Controller {
	handle := native.MaaTaskerGetController(t.handle)
	return &Controller{handle: handle}
}

// ClearCache clears runtime cache.
func (t *Tasker) ClearCache() bool {
	return native.MaaTaskerClearCache(t.handle)
}

type RecognitionDetail struct {
	ID         int64
	Name       string
	Algorithm  string
	Hit        bool
	Box        Rect
	DetailJson string
	Raw        image.Image
	Draws      []image.Image
}

// getRecognitionDetail queries recognition detail.
func (t *Tasker) getRecognitionDetail(recId int64) *RecognitionDetail {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	algorithm := buffer.NewStringBuffer()
	defer algorithm.Destroy()
	var hit bool
	box := buffer.NewRectBuffer()
	defer box.Destroy()
	detailJson := buffer.NewStringBuffer()
	defer detailJson.Destroy()
	raw := buffer.NewImageBuffer()
	defer raw.Destroy()
	draws := buffer.NewImageListBuffer()
	defer draws.Destroy()
	got := native.MaaTaskerGetRecognitionDetail(
		t.handle,
		recId,
		name.Handle(),
		algorithm.Handle(),
		&hit,
		box.Handle(),
		detailJson.Handle(),
		raw.Handle(),
		draws.Handle(),
	)
	if !got {
		return nil
	}

	rawImg := raw.Get()
	DrawImages := draws.GetAll()

	return &RecognitionDetail{
		ID:         recId,
		Name:       name.Get(),
		Algorithm:  algorithm.Get(),
		Hit:        hit,
		Box:        box.Get(),
		DetailJson: detailJson.Get(),
		Raw:        rawImg,
		Draws:      DrawImages,
	}
}

type ActionDetail struct {
	ID         int64
	Name       string
	Action     string
	Box        Rect
	Success    bool
	DetailJson string
}

func (t *Tasker) getActionDetail(actionId int64) *ActionDetail {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	action := buffer.NewStringBuffer()
	defer action.Destroy()
	box := buffer.NewRectBuffer()
	defer box.Destroy()
	var success bool
	detailJson := buffer.NewStringBuffer()
	defer detailJson.Destroy()
	got := native.MaaTaskerGetActionDetail(
		t.handle,
		actionId,
		name.Handle(),
		action.Handle(),
		box.Handle(),
		&success,
		detailJson.Handle(),
	)

	if !got {
		return nil
	}

	return &ActionDetail{
		ID:         actionId,
		Name:       name.Get(),
		Action:     action.Get(),
		Box:        box.Get(),
		Success:    success,
		DetailJson: detailJson.Get(),
	}
}

type NodeDetail struct {
	ID           int64
	Name         string
	Recognition  *RecognitionDetail
	Action       *ActionDetail
	RunCompleted bool
}

// getNodeDetail queries running detail.
func (t *Tasker) getNodeDetail(nodeId int64) *NodeDetail {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId, actionId int64
	var runCompleted bool
	got := native.MaaTaskerGetNodeDetail(
		t.handle,
		nodeId,
		name.Handle(),
		&recId,
		&actionId,
		&runCompleted,
	)
	if !got {
		return nil
	}

	recognitionDetail := t.getRecognitionDetail(recId)
	if recognitionDetail == nil {
		return nil
	}

	actionDetail := t.getActionDetail(actionId)
	if actionDetail == nil {
		return nil
	}

	return &NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
		Action:       actionDetail,
		RunCompleted: runCompleted,
	}
}

type TaskDetail struct {
	ID          int64
	Entry       string
	NodeDetails []*NodeDetail
	Status      Status
}

// getTaskDetail queries task detail.
func (t *Tasker) getTaskDetail(taskId int64) *TaskDetail {
	entry := buffer.NewStringBuffer()
	defer entry.Destroy()
	var size uint64
	got := native.MaaTaskerGetTaskDetail(
		t.handle,
		taskId,
		0,
		0,
		&size,
		nil,
	)
	if !got {
		return nil
	}
	if size == 0 {
		return &TaskDetail{
			ID:          taskId,
			Entry:       entry.Get(),
			NodeDetails: nil,
		}
	}
	nodeIdList := make([]int64, size)
	var status Status
	got = native.MaaTaskerGetTaskDetail(
		t.handle,
		taskId,
		uintptr(entry.Handle()),
		uintptr(unsafe.Pointer(&nodeIdList[0])),
		&size,
		(*int32)(&status),
	)
	if !got {
		return nil
	}

	nodeDetails := make([]*NodeDetail, size)
	for i, nodeId := range nodeIdList {
		nodeDetail := t.getNodeDetail(nodeId)
		if nodeDetail != nil {
			nodeDetails[i] = nil
		}
		nodeDetails[i] = nodeDetail
	}

	return &TaskDetail{
		ID:          taskId,
		Entry:       entry.Get(),
		NodeDetails: nodeDetails,
		Status:      status,
	}
}

// GetLatestNode returns latest node id.
func (t *Tasker) GetLatestNode(taskName string) *NodeDetail {
	var nodeId int64

	got := native.MaaTaskerGetLatestNode(t.handle, taskName, &nodeId)
	if !got {
		return nil
	}
	return t.getNodeDetail(nodeId)
}

// AddSink adds a event callback sink and returns the sink ID.
// The sink ID can be used to remove the sink later.
func (t *Tasker) AddSink(sink TaskerEventSink) int64 {
	id := registerEventCallback(sink)
	sinkId := native.MaaTaskerAddSink(
		t.handle,
		_MaaEventCallbackAgent,
		uintptr(id),
	)

	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		v.SinkIDToEventCallbackID[sinkId] = id
	})

	return sinkId
}

// RemoveSink removes a event callback sink by sink ID.
func (t *Tasker) RemoveSink(sinkId int64) {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		unregisterEventCallback(v.SinkIDToEventCallbackID[sinkId])
		delete(v.SinkIDToEventCallbackID, sinkId)
	})

	native.MaaTaskerRemoveSink(t.handle, sinkId)
}

// ClearSinks clears all event callback sinks.
func (t *Tasker) ClearSinks() {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		for _, id := range v.SinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.SinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaTaskerClearSinks(t.handle)
}

// AddContextSink adds a context event callback sink and returns the sink ID.
func (t *Tasker) AddContextSink(sink ContextEventSink) int64 {
	id := registerEventCallback(sink)
	sinkId := native.MaaTaskerAddContextSink(
		t.handle,
		_MaaEventCallbackAgent,
		uintptr(id),
	)

	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		v.ContextSinkIDToEventCallbackID[sinkId] = id
	})

	return sinkId
}

// RemoveContextSink removes a context event callback sink by sink ID.
func (t *Tasker) RemoveContextSink(sinkId int64) {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		unregisterEventCallback(v.ContextSinkIDToEventCallbackID[sinkId])
		delete(v.ContextSinkIDToEventCallbackID, sinkId)
	})

	native.MaaTaskerRemoveContextSink(t.handle, sinkId)
}

// ClearContextSinks clears all context event callback sinks.
func (t *Tasker) ClearContextSinks() {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		for _, id := range v.ContextSinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.ContextSinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaTaskerClearContextSinks(t.handle)
}
