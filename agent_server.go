package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// AgentServerRegisterCustomRecognition registers a custom recognition runner with the given name.
func AgentServerRegisterCustomRecognition(name string, recognition CustomRecognitionRunner) bool {
	id := registerCustomRecognition(recognition)

	return native.MaaAgentServerRegisterCustomRecognition(
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerRegisterCustomAction registers a custom action runner with the given name.
func AgentServerRegisterCustomAction(name string, action CustomActionRunner) bool {
	id := registerCustomAction(action)

	return native.MaaAgentServerRegisterCustomAction(
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerAddResourceSink adds a resource event callback sink and returns the sink ID.
func AgentServerAddResourceSink(sink ResourceEventSink) int64 {
	id := registerEventCallback(sink)

	return native.MaaAgentServerAddResourceSink(
		_MaaEventCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerAddControllerSink adds a controller event callback sink and returns the sink ID.
func AgentServerAddControllerSink(sink ControllerEventSink) int64 {
	id := registerEventCallback(sink)

	return native.MaaAgentServerAddControllerSink(
		_MaaEventCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerAddTaskerSink adds a tasker event callback sink and returns the sink ID.
func AgentServerAddTaskerSink(sink TaskerEventSink) int64 {
	id := registerEventCallback(sink)

	return native.MaaAgentServerAddTaskerSink(
		_MaaEventCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerAddContextSink adds a context event callback sink and returns the sink ID.
func AgentServerAddContextSink(sink ContextEventSink) int64 {
	id := registerEventCallback(sink)

	return native.MaaAgentServerAddContextSink(
		_MaaEventCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// AgentServerStartUp starts the MAA Agent Server with the given identifier.
func AgentServerStartUp(identifier string) bool {
	return native.MaaAgentServerStartUp(identifier)
}

// AgentServerShutDown shuts down the MAA Agent Server.
func AgentServerShutDown() {
	native.MaaAgentServerShutDown()
}

// AgentServerJoin waits synchronously for the service thread to finish
func AgentServerJoin() {
	native.MaaAgentServerJoin()
}

// AgentServerDetach detaches the service thread to run independently
func AgentServerDetach() {
	native.MaaAgentServerDetach()
}
