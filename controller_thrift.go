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

	agent := &callbackAgent{callback: callback}
	handle := C.MaaThriftControllerCreate(
		C.int32_t(thriftCtrlType),
		cHost,
		C.int32_t(port),
		cConfig,
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(agent)),
	)
	return &controller{handle: handle}
}
