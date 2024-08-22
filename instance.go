package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(_GoString_ msg, _GoString_ detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"errors"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"time"
	"unsafe"
)

type Instance struct {
	handle C.MaaInstanceHandle
}

// New creates an instance.
func New(callback func(msg, detailsJson string)) *Instance {
	id := registerCallback(callback)
	handle := C.MaaCreate(
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	)
	if handle == nil {
		return nil
	}
	return &Instance{handle: handle}
}

// Destroy free the instance.
func (i *Instance) Destroy() {
	C.MaaDestroy(i.handle)
}

// Handle returns the instance handle.
func (i *Instance) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}

// BindResource binds the instance to an initialized resource.
func (i *Instance) BindResource(res *Resource) bool {
	return C.MaaBindResource(i.handle, res.handle) != 0
}

// BindController binds the instance to an initialized controller.
func (i *Instance) BindController(ctrl Controller) bool {
	return C.MaaBindController(i.handle, C.MaaControllerHandle(ctrl.Handle())) != 0
}

// Inited checks if the instance is initialized.
func (i *Instance) Inited() bool {
	return C.MaaInited(i.handle) != 0
}

// RegisterCustomRecognizer registers a custom recognizer to the instance.
func (i *Instance) RegisterCustomRecognizer(name string, recognizer CustomRecognizer) bool {
	id := registerCustomRecognizer(name, recognizer)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaRegisterCustomRecognizer(
		i.handle,
		cName,
		C.MaaCustomRecognizerHandle(recognizer.Handle()),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	) != 0
}

// UnregisterCustomRecognizer unregisters a custom recognizer from the instance.
func (i *Instance) UnregisterCustomRecognizer(name string) bool {
	ok := unregisterCustomRecognizer(name)
	if !ok {
		return false
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaUnregisterCustomRecognizer(i.handle, cName) != 0
}

// ClearCustomRecognizer clears all custom recognizers registered to the instance.
func (i *Instance) ClearCustomRecognizer() bool {
	clearCustomRecognizer()
	return C.MaaClearCustomRecognizer(i.handle) != 0
}

// RegisterCustomAction registers a custom action to the instance.
func (i *Instance) RegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(name, action)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaRegisterCustomAction(
		i.handle,
		cName,
		C.MaaCustomActionHandle(action.Handle()),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	) != 0
}

// UnregisterCustomAction unregisters a custom action from the instance.
func (i *Instance) UnregisterCustomAction(name string) bool {
	ok := unregisterCustomAction(name)
	if !ok {
		return false
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaUnregisterCustomAction(i.handle, cName) != 0
}

// ClearCustomAction clears all custom actions registered to the instance.
func (i *Instance) ClearCustomAction() bool {
	clearCustomAction()
	return C.MaaClearCustomAction(i.handle) != 0
}

// PostTask posts a task to the instance.
func (i *Instance) PostTask(entry, param string) TaskJob {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	id := int64(C.MaaPostTask(i.handle, cEntry, cParam))
	return NewTaskJob(id, i.taskStatus, i.setTaskParam)
}

// PostRecognition posts a recognition to the instance.
func (i *Instance) PostRecognition(entry, param string) TaskJob {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	id := int64(C.MaaPostRecognition(i.handle, cEntry, cParam))
	return NewTaskJob(id, i.taskStatus, i.setTaskParam)
}

// PostAction posts an action to the instance.
func (i *Instance) PostAction(entry, param string) TaskJob {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	id := int64(C.MaaPostAction(i.handle, cEntry, cParam))
	return NewTaskJob(id, i.taskStatus, i.setTaskParam)
}

// setTaskParam sets the parameter of a task.
func (i *Instance) setTaskParam(id int64, param string) bool {
	cParam := C.CString(param)
	defer C.free(unsafe.Pointer(cParam))
	return C.MaaSetTaskParam(i.handle, C.int64_t(id), cParam) != 0
}

// taskStatus returns the status of a task identified by the id.
func (i *Instance) taskStatus(id int64) Status {
	return Status(C.MaaTaskStatus(i.handle, C.int64_t(id)))
}

// WaitAll waits for all tasks to complete.
func (i *Instance) WaitAll() {
	for i.Running() {
		time.Sleep(time.Millisecond * 10)
	}
}

// Running checks if the instance running.
func (i *Instance) Running() bool {
	return C.MaaRunning(i.handle) != 0
}

// PostStop posts a stop signal to the instance.
func (i *Instance) PostStop() bool {
	return C.MaaPostStop(i.handle) != 0
}

// GetResource returns the resource handle of the instance.
func (i *Instance) GetResource() *Resource {
	handle := C.MaaGetResource(i.handle)
	return &Resource{handle: handle}
}

// GetController returns the controller handle of the instance.
func (i *Instance) GetController() Controller {
	handle := C.MaaGetController(i.handle)
	return &controller{handle: handle}
}

type RecognitionDetail struct {
	ID         int64
	Name       string
	Hit        bool
	DetailJson string
	Raw        image.Image
	Draws      []image.Image
}

// QueryRecognitionDetail queries recognition detail.
func QueryRecognitionDetail(recId int64) (RecognitionDetail, error) {
	name := buffer.NewStringBuffer()
	var hit uint8
	hitBox := buffer.NewRectBuffer()
	detailJson := buffer.NewStringBuffer()
	raw := buffer.NewImageBuffer()
	draws := buffer.NewImageListBuffer()
	defer func() {
		name.Destroy()
		detailJson.Destroy()
	}()
	got := C.MaaQueryRecognitionDetail(
		C.int64_t(recId),
		C.MaaStringBufferHandle(name.Handle()),
		(*C.uint8_t)(unsafe.Pointer(&hit)),
		C.MaaRectHandle(hitBox.Handle()),
		C.MaaStringBufferHandle(detailJson.Handle()),
		C.MaaImageBufferHandle(raw.Handle()),
		C.MaaImageListBufferHandle(draws.Handle()),
	) != 0

	if !got {
		return RecognitionDetail{}, errors.New("failed to query recognition detail")
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
	RunCompleted bool
}

// QueryNodeDetail queries running detail.
func QueryNodeDetail(nodeId int64) (NodeDetail, bool) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId int64
	var runCompleted uint8
	got := C.MaaQueryNodeDetail(
		C.int64_t(nodeId),
		C.MaaStringBufferHandle(name.Handle()),
		(*C.int64_t)(unsafe.Pointer(&recId)),
		(*C.uint8_t)(unsafe.Pointer(&runCompleted)),
	) != 0

	recognitionDetail, err := QueryRecognitionDetail(recId)
	if err != nil {

	}

	return NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
		RunCompleted: runCompleted != 0,
	}, got
}

type TaskDetail struct {
	ID          int64
	Entry       string
	NodeDetails []NodeDetail
}

// QueryTaskDetail queries task detail.
func QueryTaskDetail(taskId int64) (TaskDetail, bool) {
	entry := buffer.NewStringBuffer()
	defer entry.Destroy()
	var size uint64
	got := C.MaaQueryTaskDetail(C.int64_t(taskId), nil, nil, (*C.uint64_t)(unsafe.Pointer(&size))) != 0
	if !got {
		return TaskDetail{}, got
	}
	nodeIdList := make([]int64, size)
	got = C.MaaQueryTaskDetail(
		C.int64_t(taskId),
		C.MaaStringBufferHandle(entry.Handle()),
		(*C.int64_t)(unsafe.Pointer(&nodeIdList[0])),
		(*C.uint64_t)(unsafe.Pointer(&size)),
	) != 0

	nodeDetails := make([]NodeDetail, size)
	for i, nodeId := range nodeIdList {
		nodeDetail, ok := QueryNodeDetail(nodeId)
		if !ok {

		}
		nodeDetails[i] = nodeDetail
	}

	return TaskDetail{
		ID:          taskId,
		Entry:       entry.Get(),
		NodeDetails: nodeDetails,
	}, got
}
