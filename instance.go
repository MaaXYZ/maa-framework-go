package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(_GoString_ msg, _GoString_ detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import (
	"unsafe"
)

type Instance struct {
	handle C.MaaInstanceHandle
}

func New(callback func(msg, detailsJson string, callbackArg interface{}), callbackArg interface{}) *Instance {
	agent := &callbackAgent{callback: callback, arg: callbackArg}
	handle := C.MaaCreate(
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(agent)),
	)
	return &Instance{handle: handle}
}

func (i *Instance) Destroy() {
	C.MaaDestroy(i.handle)
}

func (i *Instance) Handle() unsafe.Pointer {
	return unsafe.Pointer(i.handle)
}

func (i *Instance) BindResource(res *Resource) bool {
	return C.MaaBindResource(i.handle, res.handle) != 0
}

func (i *Instance) BindController(ctrl Controller) bool {
	return C.MaaBindController(i.handle, C.MaaControllerHandle(ctrl.Handle())) != 0
}

func (i *Instance) Inited() bool {
	return C.MaaInited(i.handle) != 0
}

func (i *Instance) RegisterCustomRecognizer(name string, recognizer CustomRecognizerImpl) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaRegisterCustomRecognizer(
		i.handle,
		cName,
		C.MaaCustomRecognizerHandle(recognizer.Handle()),
		C.MaaTransparentArg(unsafe.Pointer(&recognizer)),
	) != 0
}

func (i *Instance) UnregisterCustomRecognizer(name string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaUnregisterCustomRecognizer(i.handle, cName) != 0
}

func (i *Instance) ClearCustomRecognizer() bool {
	return C.MaaClearCustomRecognizer(i.handle) != 0
}

func (i *Instance) RegisterCustomAction(name string, action CustomActionImpl) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaRegisterCustomAction(i.handle, cName, C.MaaCustomActionHandle(action.Handle()), C.MaaTransparentArg(unsafe.Pointer(&action))) != 0
}

func (i *Instance) UnregisterCustomAction(name string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.MaaUnregisterCustomAction(i.handle, cName) != 0
}

func (i *Instance) ClearCustomAction() bool {
	return C.MaaClearCustomAction(i.handle) != 0
}

func (i *Instance) PostTask(entry, param string) int64 {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	return int64(C.MaaPostTask(i.handle, cEntry, cParam))
}

func (i *Instance) PostRecognition(entry, param string) int64 {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	return int64(C.MaaPostRecognition(i.handle, cEntry, cParam))
}

func (i *Instance) PostAction(entry, param string) int64 {
	cEntry := C.CString(entry)
	cParam := C.CString(param)
	defer func() {
		C.free(unsafe.Pointer(cEntry))
		C.free(unsafe.Pointer(cParam))
	}()
	return int64(C.MaaPostAction(i.handle, cEntry, cParam))
}

func (i *Instance) SetTaskParam(id int64, param string) bool {
	cParam := C.CString(param)
	defer C.free(unsafe.Pointer(cParam))
	return C.MaaSetTaskParam(i.handle, C.int64_t(id), cParam) != 0
}

func (i *Instance) TaskStatus(id int64) Status {
	return Status(C.MaaTaskStatus(i.handle, C.int64_t(id)))
}

func (i *Instance) WaitTask(id int64) Status {
	return Status(C.MaaWaitTask(i.handle, C.int64_t(id)))
}

func (i *Instance) Running() bool {
	return C.MaaRunning(i.handle) != 0
}

func (i *Instance) PostStop() bool {
	return C.MaaPostStop(i.handle) != 0
}

func (i *Instance) GetResource() *Resource {
	handle := C.MaaGetResource(i.handle)
	return &Resource{handle: handle}
}

func (i *Instance) GetController() Controller {
	handle := C.MaaGetController(i.handle)
	return &controller{handle: handle}
}
