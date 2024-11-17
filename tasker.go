package maa

import (
	"image"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/internal/maa"
	"github.com/MaaXYZ/maa-framework-go/internal/store"
)

var taskerStore = store.New[uint64]()

type Tasker struct {
	handle uintptr
}

// NewTasker creates an new tasker.
func NewTasker(notify Notification) *Tasker {
	id := registerNotificationCallback(notify)
	handle := maa.MaaTaskerCreate(
		_MaaNotificationCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if handle == 0 {
		return nil
	}
	taskerStore.Set(handle, id)
	return &Tasker{handle: handle}
}

// Destroy free the tasker.
func (t *Tasker) Destroy() {
	id := taskerStore.Get(t.handle)
	unregisterNotificationCallback(id)
	taskerStore.Del(t.handle)
	maa.MaaTaskerDestroy(t.handle)
}

// BindResource binds the tasker to an initialized resource.
func (t *Tasker) BindResource(res *Resource) bool {
	return maa.MaaTaskerBindResource(t.handle, res.handle)
}

// BindController binds the tasker to an initialized controller.
func (t *Tasker) BindController(ctrl Controller) bool {
	return maa.MaaTaskerBindController(t.handle, ctrl.Handle())
}

// Initialized checks if the tasker is initialized.
func (t *Tasker) Initialized() bool {
	return maa.MaaTaskerInited(t.handle)
}

func (t *Tasker) handleOverride(entry string, postFunc func(entry, override string) *TaskJob, override ...any) *TaskJob {
	if len(override) == 0 {
		return postFunc(entry, "{}")
	}
	if str, ok := override[0].(string); ok {
		return postFunc(entry, str)
	}
	str, err := toJSON(override[0])
	if err != nil {
		str = "{}"
	}
	return postFunc(entry, str)
}

func (t *Tasker) postPipeline(entry, pipelineOverride string) *TaskJob {
	id := maa.MaaTaskerPostPipeline(t.handle, entry, pipelineOverride)
	return NewTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// PostPipeline posts a task to the tasker.
// `override` is an optional parameter. If provided, it should be a single value
// that can be a JSON string or any data type that can be marshaled to JSON.
// If multiple values are provided, only the first one will be used.
func (t *Tasker) PostPipeline(entry string, override ...any) *TaskJob {
	return t.handleOverride(entry, t.postPipeline, override...)
}

// status returns the status of a task identified by the id.
func (t *Tasker) status(id int64) Status {
	return Status(maa.MaaTaskerStatus(t.handle, id))
}

// wait waits until the task is complete and returns the status of the completed task identified by the id.
func (t *Tasker) wait(id int64) Status {
	return Status(maa.MaaTaskerWait(t.handle, id))
}

// Running checks if the instance running.
func (t *Tasker) Running() bool {
	return maa.MaaTaskerRunning(t.handle)
}

// PostStop posts a stop signal to the tasker.
func (t *Tasker) PostStop() bool {
	return maa.MaaTaskerPostStop(t.handle)
}

// GetResource returns the resource handle of the tasker.
func (t *Tasker) GetResource() *Resource {
	handle := maa.MaaTaskerGetResource(t.handle)
	return &Resource{handle: handle}
}

// GetController returns the controller handle of the tasker.
func (t *Tasker) GetController() Controller {
	handle := maa.MaaTaskerGetController(t.handle)
	return &controller{handle: handle}
}

// ClearCache clears runtime cache.
func (t *Tasker) ClearCache() bool {
	return maa.MaaTaskerClearCache(t.handle)
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
	got := maa.MaaTaskerGetRecognitionDetail(
		t.handle,
		recId,
		uintptr(name.Handle()),
		uintptr(algorithm.Handle()),
		&hit,
		uintptr(box.Handle()),
		uintptr(detailJson.Handle()),
		uintptr(raw.Handle()),
		uintptr(draws.Handle()),
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

type NodeDetail struct {
	ID           int64
	Name         string
	Recognition  *RecognitionDetail
	RunCompleted bool
}

// getNodeDetail queries running detail.
func (t *Tasker) getNodeDetail(nodeId int64) *NodeDetail {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId int64
	var runCompleted bool
	got := maa.MaaTaskerGetNodeDetail(
		t.handle,
		nodeId,
		uintptr(name.Handle()),
		&recId,
		&runCompleted,
	)
	if !got {
		return nil
	}

	recognitionDetail := t.getRecognitionDetail(recId)
	if recognitionDetail == nil {
		return nil
	}

	return &NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
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
	got := maa.MaaTaskerGetTaskDetail(
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
	got = maa.MaaTaskerGetTaskDetail(
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

	got := maa.MaaTaskerGetLatestNode(t.handle, taskName, &nodeId)
	if !got {
		return nil
	}
	return t.getNodeDetail(nodeId)
}
