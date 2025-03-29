package maa

import (
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
	"unsafe"
)

func AgentServerRegisterCustomRecognition(name string, recognition CustomRecognition) bool {
	id := registerCustomRecognition(recognition)

	return maa.MaaAgentServerRegisterCustomRecognition(
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

func AgentServerRegisterCustomAction(name string, action CustomAction) bool {
	id := registerCustomAction(action)

	return maa.MaaAgentServerRegisterCustomAction(
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

func AgentServerStartUp(identifier string) bool {
	return maa.MaaAgentServerStartUp(identifier)
}

func AgentServerShutDown() {
	maa.MaaAgentServerShutDown()
}

func AgentServerJoin() {
	maa.MaaAgentServerJoin()
}

func AgentServerDetach() {
	maa.MaaAgentServerDetach()
}
