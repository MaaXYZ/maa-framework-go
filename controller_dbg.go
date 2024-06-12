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

type DbgControllerType int32

// DbgControllerType
const (
	DbgControllerTypeInvalid DbgControllerType = iota
	DbgControllerTypeCarouselImage
	DbgControllerTypeReplayRecording
)

func NewDbgController(
	readPath, writePath string,
	dbgCtrlType DbgControllerType,
	config string,
	callback func(msg, detailsJson string, callbackArg interface{}),
	callbackArg interface{},
) Controller {
	cReadPath := C.CString(readPath)
	cWritePath := C.CString(writePath)
	cConfig := C.CString(config)
	defer func() {
		C.free(unsafe.Pointer(cReadPath))
		C.free(unsafe.Pointer(cWritePath))
		C.free(unsafe.Pointer(cConfig))
	}()

	agent := &callbackAgent{callback, callbackArg}
	handle := C.MaaDbgControllerCreate(
		cReadPath,
		cWritePath,
		C.int32_t(dbgCtrlType),
		cConfig,
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(agent)),
	)
	return &controller{handle: handle}
}
