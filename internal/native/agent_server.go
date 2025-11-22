package native

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var maaAgentServer uintptr

var (
	MaaAgentServerRegisterCustomRecognition func(name string, recognition MaaCustomRecognitionCallback, transArg unsafe.Pointer) bool
	MaaAgentServerRegisterCustomAction      func(name string, action MaaCustomActionCallback, transArg unsafe.Pointer) bool
	MaaAgentServerAddResourceSink           func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddControllerSink         func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddTaskerSink             func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddContextSink            func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerStartUp                   func(identifier string) bool
	MaaAgentServerShutDown                  func()
	MaaAgentServerJoin                      func()
	MaaAgentServerDetach                    func()
)

func initAgentServer(libDir string) error {
	libName := getMaaAgentServerLibrary()
	libPath := filepath.Join(libDir, libName)

	handle, err := openLibrary(libPath)
	if err != nil {
		return err
	}

	maaAgentServer = handle

	registerAgentServer()

	return nil
}

func getMaaAgentServerLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaAgentServer.dylib"
	case "linux":
		return "libMaaAgentServer.so"
	case "windows":
		return "MaaAgentServer.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func registerAgentServer() {
	purego.RegisterLibFunc(&MaaAgentServerRegisterCustomRecognition, maaAgentServer, "MaaAgentServerRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaAgentServerRegisterCustomAction, maaAgentServer, "MaaAgentServerRegisterCustomAction")
	purego.RegisterLibFunc(&MaaAgentServerAddResourceSink, maaAgentServer, "MaaAgentServerAddResourceSink")
	purego.RegisterLibFunc(&MaaAgentServerAddControllerSink, maaAgentServer, "MaaAgentServerAddControllerSink")
	purego.RegisterLibFunc(&MaaAgentServerAddTaskerSink, maaAgentServer, "MaaAgentServerAddTaskerSink")
	purego.RegisterLibFunc(&MaaAgentServerAddContextSink, maaAgentServer, "MaaAgentServerAddContextSink")
	purego.RegisterLibFunc(&MaaAgentServerStartUp, maaAgentServer, "MaaAgentServerStartUp")
	purego.RegisterLibFunc(&MaaAgentServerShutDown, maaAgentServer, "MaaAgentServerShutDown")
	purego.RegisterLibFunc(&MaaAgentServerJoin, maaAgentServer, "MaaAgentServerJoin")
	purego.RegisterLibFunc(&MaaAgentServerDetach, maaAgentServer, "MaaAgentServerDetach")
}

func unregisterAgentServer() error {
	return unloadLibrary(maaAgentServer)
}
