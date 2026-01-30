package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// AgentServerRegisterCustomRecognition registers a custom recognition runner.
// The name should match the custom_recognition field in Pipeline.
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

// AgentServerRegisterCustomAction registers a custom action runner.
// The name should match the custom_action field in Pipeline.
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
// The identifier is used to match with AgentClient.
func AgentServerStartUp(identifier string) bool {
	return native.MaaAgentServerStartUp(identifier)
}

// AgentServerShutDown shuts down the MAA Agent Server.
func AgentServerShutDown() {
	native.MaaAgentServerShutDown()
}

// AgentServerJoin waits for the agent service to end.
// It blocks the current goroutine until the service ends.
func AgentServerJoin() {
	native.MaaAgentServerJoin()
}

// AgentServerDetach detaches the service thread to run independently.
// It allows the service to run in the background without blocking.
func AgentServerDetach() {
	native.MaaAgentServerDetach()
}
