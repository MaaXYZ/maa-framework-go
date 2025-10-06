package maa

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ebitengine/purego"
)

var maaAgentClient uintptr

var (
	MaaAgentClientCreateV2     func(identifier uintptr) uintptr
	MaaAgentClientDestroy      func(client uintptr)
	MaaAgentClientIdentifier   func(client uintptr, identifier uintptr) bool
	MaaAgentClientBindResource func(client uintptr, res uintptr) bool
	MaaAgentClientConnect      func(client uintptr) bool
	MaaAgentClientDisconnect   func(client uintptr) bool
	MaaAgentClientConnected    func(client uintptr) bool
	MaaAgentClientAlive        func(client uintptr) bool
	MaaAgentClientSetTimeout   func(client uintptr, milliseconds int64) bool
)

func initClient(libDir string) error {
	libName := getMaaAgentClientLibrary()
	libPath := filepath.Join(libDir, libName)

	handle, err := openLibrary(libPath)
	if err != nil {
		return err
	}

	maaAgentClient = handle

	registerClient()

	return nil
}

func getMaaAgentClientLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaAgentClient.dylib"
	case "linux":
		return "libMaaAgentClient.so"
	case "windows":
		return "MaaAgentClient.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func registerClient() {
	purego.RegisterLibFunc(&MaaAgentClientCreateV2, maaAgentClient, "MaaAgentClientCreateV2")
	purego.RegisterLibFunc(&MaaAgentClientDestroy, maaAgentClient, "MaaAgentClientDestroy")
	purego.RegisterLibFunc(&MaaAgentClientIdentifier, maaAgentClient, "MaaAgentClientIdentifier")
	purego.RegisterLibFunc(&MaaAgentClientBindResource, maaAgentClient, "MaaAgentClientBindResource")
	purego.RegisterLibFunc(&MaaAgentClientConnect, maaAgentClient, "MaaAgentClientConnect")
	purego.RegisterLibFunc(&MaaAgentClientDisconnect, maaAgentClient, "MaaAgentClientDisconnect")
	purego.RegisterLibFunc(&MaaAgentClientConnected, maaAgentClient, "MaaAgentClientConnected")
	purego.RegisterLibFunc(&MaaAgentClientAlive, maaAgentClient, "MaaAgentClientAlive")
	purego.RegisterLibFunc(&MaaAgentClientSetTimeout, maaAgentClient, "MaaAgentClientSetTimeout")
}
