package maa

import (
	"github.com/ebitengine/purego"
)

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

func init() {
	maaAgentClient, err := openLibrary(getMaaAgentClientLibrary())
	if err != nil {
		panic(err)
	}

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
