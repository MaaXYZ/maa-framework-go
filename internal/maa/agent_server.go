package maa

import "github.com/ebitengine/purego"

var (
	MaaAgentServerRegisterCustomRecognition func(name string, recognition MaaCustomRecognitionCallback, transArg uintptr) bool
	MaaAgentServerRegisterCustomAction      func(name string, action MaaCustomActionCallback, transArg uintptr) bool
	MaaAgentServerStartUp                   func(identifier string) bool
	MaaAgentServerShutDown                  func()
	MaaAgentServerJoin                      func()
	MaaAgentServerDetach                    func()
)

func init() {
	maaAgentServer, err := openLibrary(getMaaAgentServerLibrary())
	if err != nil {
		panic(err)
	}

	purego.RegisterLibFunc(&MaaAgentServerRegisterCustomRecognition, maaAgentServer, "MaaAgentServerRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaAgentServerRegisterCustomAction, maaAgentServer, "MaaAgentServerRegisterCustomAction")
	purego.RegisterLibFunc(&MaaAgentServerStartUp, maaAgentServer, "MaaAgentServerStartUp")
	purego.RegisterLibFunc(&MaaAgentServerShutDown, maaAgentServer, "MaaAgentServerShutDown")
	purego.RegisterLibFunc(&MaaAgentServerJoin, maaAgentServer, "MaaAgentServerJoin")
	purego.RegisterLibFunc(&MaaAgentServerDetach, maaAgentServer, "MaaAgentServerDetach")
}
