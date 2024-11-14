package maa

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	MaaToolkitProjectInterfaceRegisterCustomRecognition func(instId uint64, name string, recognition MaaCustomRecognitionCallback, transArg unsafe.Pointer)
	MaaToolkitProjectInterfaceRegisterCustomAction      func(instId uint64, name string, action MaaCustomActionCallback, transArg unsafe.Pointer)
	MaaToolkitProjectInterfaceRunCli                    func(instId uint64, resourcePath, userPath string, directly bool, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) bool
)

func init() {
	maaToolkit, err := openLibrary(getMaaToolkitLibrary())
	if err != nil {
		panic(err)
	}
	// ProjectInterface
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRegisterCustomRecognition, maaToolkit, "MaaToolkitProjectInterfaceRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRegisterCustomAction, maaToolkit, "MaaToolkitProjectInterfaceRegisterCustomAction")
	purego.RegisterLibFunc(&MaaToolkitProjectInterfaceRunCli, maaToolkit, "MaaToolkitProjectInterfaceRunCli")
}
