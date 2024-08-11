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

type AdbControllerType int32

// AdbControllerType
const (
	AdbControllerTypeInvalid AdbControllerType = iota

	AdbControllerTypeTouchAdb
	AdbControllerTypeTouchMiniTouch
	AdbControllerTypeTouchMaaTouch
	AdbControllerTypeTouchEmulatorExtras
	AdbControllerTypeTouchAutoDetect AdbControllerType = 0xFF - 1

	AdbControllerTypeKeyAdb            AdbControllerType = 1 << 8
	AdbControllerTypeKeyMaaTouch       AdbControllerType = 2 << 8
	AdbControllerTypeKeyEmulatorExtras AdbControllerType = 3 << 8
	AdbControllerTypeKeyAutoDetect     AdbControllerType = 0xFF00 - (1 << 8)

	AdbControllerTypeInputPresetAdb            = AdbControllerTypeTouchAdb | AdbControllerTypeKeyAdb
	AdbControllerTypeInputPresetMiniTouch      = AdbControllerTypeTouchMiniTouch | AdbControllerTypeKeyAdb
	AdbControllerTypeInputPresetMaaTouch       = AdbControllerTypeTouchMaaTouch | AdbControllerTypeKeyMaaTouch
	AdbControllerTypeInputPresetEmulatorExtras = AdbControllerTypeTouchEmulatorExtras | AdbControllerTypeKeyEmulatorExtras
	AdbControllerTypeInputPresetAutoDetect     = AdbControllerTypeTouchAutoDetect | AdbControllerTypeKeyAutoDetect

	AdbControllerTypeScreencapFastestWayCompatible AdbControllerType = 1 << 16
	AdbControllerTypeScreencapRawByNetcat          AdbControllerType = 2 << 16
	AdbControllerTypeScreencapRawWithGzip          AdbControllerType = 3 << 16
	AdbControllerTypeScreencapEncode               AdbControllerType = 4 << 16
	AdbControllerTypeScreencapEncodeToFile         AdbControllerType = 5 << 16
	AdbControllerTypeScreencapMinicapDirect        AdbControllerType = 6 << 16
	AdbControllerTypeScreencapMinicapStream        AdbControllerType = 7 << 16
	AdbControllerTypeScreencapEmulatorExtras       AdbControllerType = 8 << 16
	AdbControllerTypeScreencapFastestLosslessWay   AdbControllerType = 0xFF0000 - (2 << 16)
	AdbControllerTypeScreencapFastestWay           AdbControllerType = 0xFF0000 - (1 << 16)
)

// NewAdbController creates an ADB controller instance.
func NewAdbController(
	adbPath, address string,
	adbCtrlType AdbControllerType,
	config, agentPath string,
	callback func(msg, detailsJson string),
) Controller {
	cAdbPath := C.CString(adbPath)
	cAddress := C.CString(address)
	cConfig := C.CString(config)
	cAgentPath := C.CString(agentPath)
	defer func() {
		C.free(unsafe.Pointer(cAdbPath))
		C.free(unsafe.Pointer(cAddress))
		C.free(unsafe.Pointer(cConfig))
		C.free(unsafe.Pointer(cAgentPath))
	}()

	agent := &callbackAgent{callback: callback}
	handle := C.MaaAdbControllerCreateV2(
		cAdbPath,
		cAddress,
		C.int32_t(adbCtrlType),
		cConfig,
		cAgentPath,
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(agent)),
	)
	return &controller{handle: handle}
}
