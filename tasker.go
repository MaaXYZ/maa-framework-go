package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);
*/
import "C"
import (
	"errors"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"time"
	"unsafe"
)

type Tasker struct {
	handle *C.MaaTasker
}

// New creates an instance.
func New(callback func(msg, detailsJson string)) *Tasker {
	id := registerNotificationCallback(callback)
	handle := C.MaaTaskerCreate(
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if handle == nil {
		return nil
	}
	return &Tasker{handle: handle}
}

// Destroy free the instance.
func (t *Tasker) Destroy() {
	C.MaaTaskerDestroy(t.handle)
}

// Handle returns the instance handle.
func (t *Tasker) Handle() unsafe.Pointer {
	return unsafe.Pointer(t.handle)
}

// BindResource binds the instance to an initialized resource.
func (t *Tasker) BindResource(res *Resource) bool {
	return C.MaaTaskerBindResource(t.handle, res.handle) != 0
}

// BindController binds the instance to an initialized controller.
func (t *Tasker) BindController(ctrl Controller) bool {
	return C.MaaTaskerBindController(t.handle, (*C.MaaController)(ctrl.Handle())) != 0
}

// Inited checks if the instance is initialized.
func (t *Tasker) Inited() bool {
	return C.MaaTaskerInited(t.handle) != 0
}

// PostPipeline posts a task to the instance.
func (t *Tasker) PostPipeline(entry, pipelineOverride string) TaskJob {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	id := int64(C.MaaTaskerPostPipeline(t.handle, cEntry, cPipelineOverride))
	return NewTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// PostRecognition posts a recognition to the instance.
func (t *Tasker) PostRecognition(entry, pipelineOverride string) TaskJob {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	id := int64(C.MaaTaskerPostRecognition(t.handle, cEntry, cPipelineOverride))
	return NewTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// PostAction posts an action to the instance.
func (t *Tasker) PostAction(entry, pipelineOverride string) TaskJob {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	id := int64(C.MaaTaskerPostAction(t.handle, cEntry, cPipelineOverride))
	return NewTaskJob(id, t.status, t.wait, t.getTaskDetail)
}

// status returns the status of a task identified by the id.
func (t *Tasker) status(id int64) Status {
	return Status(C.MaaTaskerStatus(t.handle, C.int64_t(id)))
}

// wait waits until the task is complete and returns the status of the completed task identified by the id.
func (t *Tasker) wait(id int64) Status {
	return Status(C.MaaTaskerWait(t.handle, C.int64_t(id)))
}

// WaitAll waits for all tasks to complete.
func (t *Tasker) WaitAll() {
	for t.Running() {
		time.Sleep(time.Millisecond * 10)
	}
}

// Running checks if the instance running.
func (t *Tasker) Running() bool {
	return C.MaaTaskerRunning(t.handle) != 0
}

// PostStop posts a stop signal to the instance.
func (t *Tasker) PostStop() bool {
	return C.MaaTaskerPostStop(t.handle) != 0
}

// GetResource returns the resource handle of the instance.
func (t *Tasker) GetResource() *Resource {
	handle := C.MaaTaskerGetResource(t.handle)
	return &Resource{handle: handle}
}

// GetController returns the controller handle of the instance.
func (t *Tasker) GetController() Controller {
	handle := C.MaaTaskerGetController(t.handle)
	return &controller{handle: handle}
}

// ClearCache clears runtime cache.
func (t *Tasker) ClearCache() bool {
	return C.MaaTaskerClearCache(t.handle) != 0
}

type RecognitionDetail struct {
	ID         int64
	Name       string
	Hit        bool
	DetailJson string
	Raw        image.Image
	Draws      []image.Image
}

// getRecognitionDetail queries recognition detail.
func (t *Tasker) getRecognitionDetail(recId int64) (RecognitionDetail, error) {
	name := buffer.NewStringBuffer()
	var hit uint8
	hitBox := newRectBuffer()
	detailJson := buffer.NewStringBuffer()
	raw := buffer.NewImageBuffer()
	draws := buffer.NewImageListBuffer()
	defer func() {
		name.Destroy()
		detailJson.Destroy()
	}()
	got := C.MaaTaskerGetRecognitionDetail(
		t.handle,
		C.int64_t(recId),
		(*C.MaaStringBuffer)(name.Handle()),
		(*C.uint8_t)(unsafe.Pointer(&hit)),
		(*C.MaaRect)(hitBox.Handle()),
		(*C.MaaStringBuffer)(detailJson.Handle()),
		(*C.MaaImageBuffer)(raw.Handle()),
		(*C.MaaImageListBuffer)(draws.Handle()),
	) != 0

	if !got {
		return RecognitionDetail{}, errors.New("failed to get recognition detail")
	}

	rawImg, err := raw.GetByRawData()
	if err != nil {
		return RecognitionDetail{}, err
	}

	DrawImages, err := draws.GetAll()
	if err != nil {
		return RecognitionDetail{}, err
	}

	return RecognitionDetail{
		ID:         recId,
		Name:       name.Get(),
		Hit:        hit != 0,
		DetailJson: detailJson.Get(),
		Raw:        rawImg,
		Draws:      DrawImages,
	}, nil
}

type NodeDetail struct {
	ID           int64
	Name         string
	Recognition  RecognitionDetail
	Times        uint64
	RunCompleted bool
}

// getNodeDetail queries running detail.
func (t *Tasker) getNodeDetail(nodeId int64) (NodeDetail, bool) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId int64
	var times uint64
	var runCompleted uint8
	got := C.MaaTaskerGetNodeDetail(
		t.handle,
		C.int64_t(nodeId),
		(*C.MaaStringBuffer)(name.Handle()),
		(*C.int64_t)(unsafe.Pointer(&recId)),
		(*C.uint64_t)(unsafe.Pointer(&times)),
		(*C.uint8_t)(unsafe.Pointer(&runCompleted)),
	)

	recognitionDetail, err := t.getRecognitionDetail(recId)
	if err != nil {
		// todo: handle error
		return NodeDetail{}, false
	}

	return NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
		Times:        times,
		RunCompleted: runCompleted != 0,
	}, got != 0
}

type TaskDetail struct {
	ID          int64
	Entry       string
	NodeDetails []NodeDetail
}

// getTaskDetail queries task detail.
func (t *Tasker) getTaskDetail(taskId int64) (TaskDetail, bool) {
	entry := buffer.NewStringBuffer()
	defer entry.Destroy()
	var size uint64
	got := C.MaaTaskerGetTaskDetail(t.handle, C.int64_t(taskId), nil, nil, (*C.uint64_t)(unsafe.Pointer(&size)))
	if got == 0 {
		return TaskDetail{}, false
	}
	nodeIdList := make([]int64, size)
	got = C.MaaTaskerGetTaskDetail(
		t.handle,
		C.int64_t(taskId),
		(*C.MaaStringBuffer)(entry.Handle()),
		(*C.int64_t)(unsafe.Pointer(&nodeIdList[0])),
		(*C.uint64_t)(unsafe.Pointer(&size)),
	)

	nodeDetails := make([]NodeDetail, size)
	for i, nodeId := range nodeIdList {
		nodeDetail, ok := t.getNodeDetail(nodeId)
		if !ok {
			nodeDetails[i] = NodeDetail{}
		}
		nodeDetails[i] = nodeDetail
	}

	return TaskDetail{
		ID:          taskId,
		Entry:       entry.Get(),
		NodeDetails: nodeDetails,
	}, got != 0
}

// GetLatestNode returns latest node id.
func (t *Tasker) GetLatestNode(taskName string) (NodeDetail, bool) {
	cTaskName := C.CString(taskName)
	defer C.free(unsafe.Pointer(cTaskName))
	var nodeId int64

	got := C.MaaTaskerGetLatestNode(t.handle, cTaskName, (*C.int64_t)(unsafe.Pointer(&nodeId)))
	if got == 0 {
		return NodeDetail{}, false
	}
	return t.getNodeDetail(nodeId)
}
