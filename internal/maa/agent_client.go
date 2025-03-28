package maa

import "github.com/ebitengine/purego"

var (
	MaaAgentClientCreate       func() uintptr
	MaaAgentClientDestroy      func(client uintptr)
	MaaAgentClientBindResource func(client uintptr, res uintptr) bool
	// if identifier is empty, bind to default address, and output the identifier. otherwise bind to the specified identifier
	MaaAgentClientCreateSocket func(client uintptr, identifier uintptr) bool
	MaaAgentClientConnect      func(client uintptr) bool
	MaaAgentClientDisconnect   func(client uintptr) bool
)

func init() {
	maaAgentClient, err := openLibrary(getMaaAgentClientLibrary())
	if err != nil {
		panic(err)
	}

	purego.RegisterLibFunc(&MaaAgentClientCreate, maaAgentClient, "MaaAgentClientCreate")
	purego.RegisterLibFunc(&MaaAgentClientDestroy, maaAgentClient, "MaaAgentClientDestroy")
	purego.RegisterLibFunc(&MaaAgentClientBindResource, maaAgentClient, "MaaAgentClientBindResource")
	purego.RegisterLibFunc(&MaaAgentClientCreateSocket, maaAgentClient, "MaaAgentClientCreateSocket")
	purego.RegisterLibFunc(&MaaAgentClientConnect, maaAgentClient, "MaaAgentClientConnect")
	purego.RegisterLibFunc(&MaaAgentClientDisconnect, maaAgentClient, "MaaAgentClientDisconnect")
}
