package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
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

// NewDbgController creates a DBG controller instance.
func NewDbgController(
	readPath, writePath string,
	dbgCtrlType DbgControllerType,
	config string,
	callback func(msg, detailsJson string),
) Controller {
	cReadPath := C.CString(readPath)
	cWritePath := C.CString(writePath)
	cConfig := C.CString(config)
	defer func() {
		C.free(unsafe.Pointer(cReadPath))
		C.free(unsafe.Pointer(cWritePath))
		C.free(unsafe.Pointer(cConfig))
	}()

	id := registerCallback(callback)
	handle := C.MaaDbgControllerCreate(
		cReadPath,
		cWritePath,
		C.int32_t(dbgCtrlType),
		cConfig,
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	)
	return &controller{handle: handle}
}
