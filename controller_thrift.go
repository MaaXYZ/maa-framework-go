package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>

extern void _MaaAPICallbackAgent(MaaStringView msg, MaaStringView detailsJson, MaaTransparentArg callbackArg);
*/
import "C"
import "unsafe"

type ThriftControllerType int32

const (
	ThriftControllerInvalid ThriftControllerType = iota
	ThriftControllerTypeSocket
	ThriftControllerTypeUnixDomainSocket
)

// NewThriftController creates a thrift controller instance.
func NewThriftController(
	thriftCtrlType ThriftControllerType,
	host string,
	port int32,
	config string,
	callback func(msg, detailsJson string),
) Controller {
	cHost := C.CString(host)
	cConfig := C.CString(config)
	defer func() {
		C.free(unsafe.Pointer(cHost))
		C.free(unsafe.Pointer(cConfig))

	}()

	id := registerCallback(callback)
	handle := C.MaaThriftControllerCreate(
		C.int32_t(thriftCtrlType),
		cHost,
		C.int32_t(port),
		cConfig,
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		C.MaaTransparentArg(unsafe.Pointer(uintptr(id))),
	)
	return &controller{handle: handle}
}
