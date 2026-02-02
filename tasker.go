package maa

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/store"
)

// Tasker is the main task executor that coordinates resources and controllers
// to perform automated tasks.
type Tasker struct {
	handle uintptr
}

// NewTasker creates a new tasker instance.
func NewTasker() (*Tasker, error) {
	handle := native.MaaTaskerCreate()
	if handle == 0 {
		return nil, errors.New("failed to create tasker")
	}

	store.TaskerStore.Lock()
	store.TaskerStore.Set(handle, store.TaskerStoreValue{
		SinkIDToEventCallbackID:        make(map[int64]uint64),
		ContextSinkIDToEventCallbackID: make(map[int64]uint64),
	})
	store.TaskerStore.Unlock()

	return &Tasker{
		handle: handle,
	}, nil
}

// Destroy frees the tasker and releases all associated resources.
// After calling this method, the tasker should not be used anymore.
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

// BindResource binds an initialized resource to the tasker.
func (t *Tasker) BindResource(res *Resource) error {
	ok := native.MaaTaskerBindResource(t.handle, res.handle)
	if !ok {
		return errors.New("failed to bind resource")
	}
	return nil
}

// BindController binds an initialized controller to the tasker.
func (t *Tasker) BindController(ctrl *Controller) error {
	ok := native.MaaTaskerBindController(t.handle, ctrl.handle)
	if !ok {
		return errors.New("failed to bind controller")
	}
	return nil
}

// Initialized checks if the tasker is correctly initialized.
// A tasker is considered initialized when both a resource and a controller are bound.
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
			return newTaskJob(0, nil, nil, nil, nil,
				fmt.Errorf("failed to marshal override: %w", err))
		}
		return postFunc(entry, string(jsonBytes))
	}
}

func (t *Tasker) postTask(entry, pipelineOverride string) *TaskJob {
	id := native.MaaTaskerPostTask(t.handle, entry, pipelineOverride)
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail, t.overridePipeline, nil)
}

// PostTask posts a task to the tasker asynchronously.
// The optional override can be a JSON string, []byte, or any JSON-marshalable value.
func (t *Tasker) PostTask(entry string, override ...any) *TaskJob {
	return t.handleOverride(entry, t.postTask, override...)
}

// PostRecognition posts a recognition to the tasker asynchronously.
func (t *Tasker) PostRecognition(recType NodeRecognitionType, recParam NodeRecognitionParam, img image.Image) *TaskJob {
	imgBuf := buffer.NewImageBuffer()
	defer imgBuf.Destroy()
	imgBuf.Set(img)

	recParamJSON, err := json.Marshal(recParam)
	if err != nil {
		return newTaskJob(0, nil, nil, nil, nil,
			fmt.Errorf("failed to marshal recognition param: %w", err))
	}

	id := native.MaaTaskerPostRecognition(t.handle, string(recType), string(recParamJSON), imgBuf.Handle())
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail, t.overridePipeline, nil)
}

// PostAction posts an action to the tasker asynchronously.
// The box and recoDetail are from the previous recognition.
func (t *Tasker) PostAction(actionType NodeActionType, actionParam NodeActionParam, box Rect, recoDetail *RecognitionDetail) *TaskJob {
	rectBuf := buffer.NewRectBuffer()
	defer rectBuf.Destroy()
	rectBuf.Set(box)

	actParamJSON, err := json.Marshal(actionParam)
	if err != nil {
		return newTaskJob(0, nil, nil, nil, nil,
			fmt.Errorf("failed to marshal action param: %w", err))
	}

	recoDetailJSON, err := json.Marshal(recoDetail)
	if err != nil {
		return newTaskJob(0, nil, nil, nil, nil,
			fmt.Errorf("failed to marshal recognition detail: %w", err))
	}

	id := native.MaaTaskerPostAction(t.handle, string(actionType), string(actParamJSON), rectBuf.Handle(), string(recoDetailJSON))
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail, t.overridePipeline, nil)
}

// Stopping checks if the tasker is in the process of stopping (not yet fully stopped).
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

// Running checks if the tasker is currently running a task.
func (t *Tasker) Running() bool {
	return native.MaaTaskerRunning(t.handle)
}

// PostStop posts a stop signal to the tasker asynchronously.
// It interrupts the currently running task and stops resource loading and controller operations.
func (t *Tasker) PostStop() *TaskJob {
	id := native.MaaTaskerPostStop(t.handle)
	return newTaskJob(id, t.status, t.wait, t.getTaskDetail, t.overridePipeline, nil)
}

// GetResource returns the bound resource of the tasker.
func (t *Tasker) GetResource() *Resource {
	handle := native.MaaTaskerGetResource(t.handle)
	return &Resource{handle: handle}
}

// GetController returns the bound controller of the tasker.
func (t *Tasker) GetController() *Controller {
	handle := native.MaaTaskerGetController(t.handle)
	return &Controller{handle: handle}
}

// ClearCache clears all queryable runtime cache.
func (t *Tasker) ClearCache() error {
	ok := native.MaaTaskerClearCache(t.handle)
	if !ok {
		return errors.New("failed to clear cache")
	}
	return nil
}

func (t *Tasker) overridePipeline(taskId int64, override any) error {
	var overrideStr string
	switch v := override.(type) {
	case string:
		overrideStr = v
	case []byte:
		overrideStr = string(v)
	default:
		if v == nil {
			overrideStr = "{}"
		} else {
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return err
			}
			overrideStr = string(jsonBytes)
		}
	}
	ok := native.MaaTaskerOverridePipeline(t.handle, taskId, overrideStr)
	if !ok {
		return errors.New("failed to override pipeline")
	}
	return nil
}

// RecognitionDetail contains recognition information.
type RecognitionDetail struct {
	ID             int64
	Name           string
	Algorithm      string
	Hit            bool
	Box            Rect
	DetailJson     string
	Results        *RecognitionResults  // nil if algorithm is DirectHit, And or Or.
	CombinedResult []*RecognitionDetail // for And/Or algorithms only.
	Raw            image.Image          // available when debug mode or save_draw is enabled.
	Draws          []image.Image        // available when debug mode or save_draw is enabled.
}

// getRecognitionDetail queries recognition detail.
func (t *Tasker) getRecognitionDetail(recId int64) (*RecognitionDetail, error) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	algorithm := buffer.NewStringBuffer()
	defer algorithm.Destroy()
	var hitByte uint8 // Use uint8 instead of bool for C ABI compatibility on macOS
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
		(*bool)(unsafe.Pointer(&hitByte)), // Convert uint8* to bool* for FFI call
		box.Handle(),
		detailJson.Handle(),
		raw.Handle(),
		draws.Handle(),
	)
	if !got {
		return nil, errors.New("failed to get recognition detail")
	}

	rawImg := raw.Get()
	DrawImages := draws.GetAll()

	detailJsonStr := detailJson.Get()
	algorithmStr := algorithm.Get()

	var err error
	var results *RecognitionResults
	var combinedResults []*RecognitionDetail

	if isCombinedRecognition(algorithmStr) {
		combinedResults, err = parseCombinedResult(detailJsonStr)
		if err != nil {
			return nil, err
		}
	} else {
		results, err = parseRecognitionResults(algorithmStr, detailJsonStr)
		if err != nil {
			return nil, err
		}
	}

	return &RecognitionDetail{
		ID:             recId,
		Name:           name.Get(),
		Algorithm:      algorithmStr,
		Hit:            hitByte != 0,
		Box:            box.Get(),
		DetailJson:     detailJsonStr,
		Results:        results,
		CombinedResult: combinedResults,
		Raw:            rawImg,
		Draws:          DrawImages,
	}, nil
}

// ActionDetail contains action information.
type ActionDetail struct {
	ID         int64
	Name       string
	Action     string
	Box        Rect
	Success    bool
	DetailJson string
	Result     *ActionResult
}

func (t *Tasker) getActionDetail(actionId int64) (*ActionDetail, error) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	action := buffer.NewStringBuffer()
	defer action.Destroy()
	box := buffer.NewRectBuffer()
	defer box.Destroy()
	var successByte uint8 // Use uint8 instead of bool for C ABI compatibility on macOS
	detailJson := buffer.NewStringBuffer()
	defer detailJson.Destroy()
	got := native.MaaTaskerGetActionDetail(
		t.handle,
		actionId,
		name.Handle(),
		action.Handle(),
		box.Handle(),
		(*bool)(unsafe.Pointer(&successByte)), // Convert uint8* to bool* for FFI call
		detailJson.Handle(),
	)

	if !got {
		return nil, errors.New("failed to get action detail")
	}

	detailJsonStr := detailJson.Get()
	result, err := parseActionResult(action.Get(), detailJsonStr)
	if err != nil {
		return nil, err
	}

	return &ActionDetail{
		ID:         actionId,
		Name:       name.Get(),
		Action:     action.Get(),
		Box:        box.Get(),
		Success:    successByte != 0,
		DetailJson: detailJsonStr,
		Result:     result,
	}, nil
}

// NodeDetail contains node information.
type NodeDetail struct {
	ID           int64
	Name         string
	Recognition  *RecognitionDetail
	Action       *ActionDetail
	RunCompleted bool
}

// getNodeDetail queries running detail.
func (t *Tasker) getNodeDetail(nodeId int64) (*NodeDetail, error) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId, actionId int64
	var runCompletedByte uint8 // Use uint8 instead of bool for C ABI compatibility on macOS
	got := native.MaaTaskerGetNodeDetail(
		t.handle,
		nodeId,
		name.Handle(),
		&recId,
		&actionId,
		(*bool)(unsafe.Pointer(&runCompletedByte)), // Convert uint8* to bool* for FFI call
	)
	if !got {
		return nil, errors.New("failed to get node detail")
	}

	recognitionDetail, err := t.getRecognitionDetail(recId)
	if err != nil {
		return nil, err
	}

	actionDetail, err := t.getActionDetail(actionId)
	if err != nil {
		return nil, err
	}

	return &NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
		Action:       actionDetail,
		RunCompleted: runCompletedByte != 0,
	}
}

// TaskDetail contains task information.
type TaskDetail struct {
	ID          int64
	Entry       string
	NodeDetails []*NodeDetail
	Status      Status
}

// getTaskDetail queries task detail.
func (t *Tasker) getTaskDetail(taskId int64) (*TaskDetail, error) {
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
		return nil, errors.New("failed to get task detail size")
	}
	if size == 0 {
		return &TaskDetail{
			ID:          taskId,
			Entry:       entry.Get(),
			NodeDetails: nil,
		}, nil
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
		return nil, errors.New("failed to get task detail data")
	}

	var err error
	nodeDetails := make([]*NodeDetail, size)
	for i, nodeId := range nodeIdList {
		nodeDetails[i], err = t.getNodeDetail(nodeId)
		if err != nil {
			return nil, err
		}
	}

	return &TaskDetail{
		ID:          taskId,
		Entry:       entry.Get(),
		NodeDetails: nodeDetails,
		Status:      status,
	}, nil
}

// GetLatestNode returns the latest node detail for a given task name.
func (t *Tasker) GetLatestNode(taskName string) (*NodeDetail, error) {
	var nodeId int64

	got := native.MaaTaskerGetLatestNode(t.handle, taskName, &nodeId)
	if !got {
		return nil, errors.New("failed to get latest node")
	}
	return t.getNodeDetail(nodeId)
}

// AddSink adds an event listener and returns the sink ID for later removal.
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

// RemoveSink removes an event listener by sink ID.
func (t *Tasker) RemoveSink(sinkId int64) {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		unregisterEventCallback(v.SinkIDToEventCallbackID[sinkId])
		delete(v.SinkIDToEventCallbackID, sinkId)
	})

	native.MaaTaskerRemoveSink(t.handle, sinkId)
}

// ClearSinks clears all instance event listeners.
func (t *Tasker) ClearSinks() {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		for _, id := range v.SinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.SinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaTaskerClearSinks(t.handle)
}

// AddContextSink adds a context event listener and returns the sink ID for later removal.
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

// RemoveContextSink removes a context event listener by sink ID.
func (t *Tasker) RemoveContextSink(sinkId int64) {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		unregisterEventCallback(v.ContextSinkIDToEventCallbackID[sinkId])
		delete(v.ContextSinkIDToEventCallbackID, sinkId)
	})

	native.MaaTaskerRemoveContextSink(t.handle, sinkId)
}

// ClearContextSinks clears all context event listeners.
func (t *Tasker) ClearContextSinks() {
	store.TaskerStore.Update(t.handle, func(v *store.TaskerStoreValue) {
		for _, id := range v.ContextSinkIDToEventCallbackID {
			unregisterEventCallback(id)
		}
		v.ContextSinkIDToEventCallbackID = make(map[int64]uint64)
	})

	native.MaaTaskerClearContextSinks(t.handle)
}

// TaskerEventSink is the interface for receiving tasker-level events.
type TaskerEventSink interface {
	OnTaskerTask(tasker *Tasker, event EventStatus, detail TaskerTaskDetail)
}

// taskerEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type taskerEventSinkAdapter struct {
	onTaskerTask func(EventStatus, TaskerTaskDetail)
}

func (a *taskerEventSinkAdapter) OnTaskerTask(tasker *Tasker, status EventStatus, detail TaskerTaskDetail) {
	if a == nil || a.onTaskerTask == nil {
		return
	}
	a.onTaskerTask(status, detail)
}

// OnTaskerTask registers a callback for Tasker.Task events and returns the sink ID.
func (t *Tasker) OnTaskerTask(fn func(EventStatus, TaskerTaskDetail)) int64 {
	sink := &taskerEventSinkAdapter{onTaskerTask: fn}
	return t.AddSink(sink)
}

// ContextEventSink is the interface for receiving context-level events.
type ContextEventSink interface {
	OnNodePipelineNode(ctx *Context, event EventStatus, detail NodePipelineNodeDetail)
	OnNodeRecognitionNode(ctx *Context, event EventStatus, detail NodeRecognitionNodeDetail)
	OnNodeActionNode(ctx *Context, event EventStatus, detail NodeActionNodeDetail)
	OnNodeNextList(ctx *Context, event EventStatus, detail NodeNextListDetail)
	OnNodeRecognition(ctx *Context, event EventStatus, detail NodeRecognitionDetail)
	OnNodeAction(ctx *Context, event EventStatus, detail NodeActionDetail)
}

// contextEventSinkAdapter is a lightweight adapter that makes it easy to register
// a single-event handler via a callback function.
type contextEventSinkAdapter struct {
	onNodePipelineNode    func(*Context, EventStatus, NodePipelineNodeDetail)
	onNodeRecognitionNode func(*Context, EventStatus, NodeRecognitionNodeDetail)
	onNodeActionNode      func(*Context, EventStatus, NodeActionNodeDetail)
	onNodeNextList        func(*Context, EventStatus, NodeNextListDetail)
	onNodeRecognition     func(*Context, EventStatus, NodeRecognitionDetail)
	onNodeAction          func(*Context, EventStatus, NodeActionDetail)
}

func (a *contextEventSinkAdapter) OnNodePipelineNode(ctx *Context, status EventStatus, detail NodePipelineNodeDetail) {
	if a == nil || a.onNodePipelineNode == nil {
		return
	}
	a.onNodePipelineNode(ctx, status, detail)
}

func (a *contextEventSinkAdapter) OnNodeRecognitionNode(ctx *Context, status EventStatus, detail NodeRecognitionNodeDetail) {
	if a == nil || a.onNodeRecognitionNode == nil {
		return
	}
	a.onNodeRecognitionNode(ctx, status, detail)
}

func (a *contextEventSinkAdapter) OnNodeActionNode(ctx *Context, status EventStatus, detail NodeActionNodeDetail) {
	if a == nil || a.onNodeActionNode == nil {
		return
	}
	a.onNodeActionNode(ctx, status, detail)
}

func (a *contextEventSinkAdapter) OnNodeNextList(ctx *Context, status EventStatus, detail NodeNextListDetail) {
	if a == nil || a.onNodeNextList == nil {
		return
	}
	a.onNodeNextList(ctx, status, detail)
}

func (a *contextEventSinkAdapter) OnNodeRecognition(ctx *Context, status EventStatus, detail NodeRecognitionDetail) {
	if a == nil || a.onNodeRecognition == nil {
		return
	}
	a.onNodeRecognition(ctx, status, detail)
}

func (a *contextEventSinkAdapter) OnNodeAction(ctx *Context, status EventStatus, detail NodeActionDetail) {
	if a == nil || a.onNodeAction == nil {
		return
	}
	a.onNodeAction(ctx, status, detail)
}

// OnNodePipelineNodeInContext registers a callback for Node.PipelineNode events and returns the sink ID.
func (t *Tasker) OnNodePipelineNodeInContext(fn func(*Context, EventStatus, NodePipelineNodeDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodePipelineNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeRecognitionNodeInContext registers a callback for Node.RecognitionNode events and returns the sink ID.
func (t *Tasker) OnNodeRecognitionNodeInContext(fn func(*Context, EventStatus, NodeRecognitionNodeDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodeRecognitionNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeActionNodeInContext registers a callback for Node.ActionNode events and returns the sink ID.
func (t *Tasker) OnNodeActionNodeInContext(fn func(*Context, EventStatus, NodeActionNodeDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodeActionNode: fn}
	return t.AddContextSink(sink)
}

// OnNodeNextListInContext registers a callback for Node.NextList events and returns the sink ID.
func (t *Tasker) OnNodeNextListInContext(fn func(*Context, EventStatus, NodeNextListDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodeNextList: fn}
	return t.AddContextSink(sink)
}

// OnNodeRecognitionInContext registers a callback for Node.Recognition events and returns the sink ID.
func (t *Tasker) OnNodeRecognitionInContext(fn func(*Context, EventStatus, NodeRecognitionDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodeRecognition: fn}
	return t.AddContextSink(sink)
}

// OnNodeActionInContext registers a callback for Node.Action events and returns the sink ID.
func (t *Tasker) OnNodeActionInContext(fn func(*Context, EventStatus, NodeActionDetail)) int64 {
	sink := &contextEventSinkAdapter{onNodeAction: fn}
	return t.AddContextSink(sink)
}
