package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

// AgentServerRegisterCustomRecognition registers a custom recognition with the agent server
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

// AgentServerRegisterCustomAction registers a custom action with the agent server
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

// AgentServerStartUp starts up the agent server with the given identifier
func AgentServerStartUp(identifier string) bool {
	return maa.MaaAgentServerStartUp(identifier)
}

// AgentServerShutDown shuts down the Maa agent server
func AgentServerShutDown() {
	maa.MaaAgentServerShutDown()
}

// AgentServerJoin registers the agent server
func AgentServerJoin() {
	maa.MaaAgentServerJoin()
}

// AgentServerDetach unregisters the agent server
func AgentServerDetach() {
	maa.MaaAgentServerDetach()
}
