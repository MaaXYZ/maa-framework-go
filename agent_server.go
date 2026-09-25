package maa

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

type agentServerPhase uint32

const (
	agentServerStopped agentServerPhase = iota
	agentServerRunningAttached
	agentServerJoined
	agentServerDetached
)

// A detached native thread cannot be joined or observed through the current
// MaaAgentServer API, so agentServerDetached is a terminal state for Release.
var agentServerState atomic.Uint32

// AgentServerRegisterCustomRecognition registers a custom recognition runner.
// The name should match the custom_recognition field in Pipeline.
func AgentServerRegisterCustomRecognition(name string, recognition CustomRecognitionRunner) error {
	id := registerCustomRecognition(recognition)

	ok := native.MaaAgentServerRegisterCustomRecognition(
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if !ok {
		unregisterCustomRecognition(id)
		return fmt.Errorf("failed to register custom recognition: %s", name)
	}
	return nil
}

// AgentServerRegisterCustomAction registers a custom action runner.
// The name should match the custom_action field in Pipeline.
func AgentServerRegisterCustomAction(name string, action CustomActionRunner) error {
	id := registerCustomAction(action)

	ok := native.MaaAgentServerRegisterCustomAction(
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	if !ok {
		unregisterCustomAction(id)
		return fmt.Errorf("failed to register custom action: %s", name)
	}
	return nil
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
func AgentServerStartUp(identifier string) error {
	if !native.MaaAgentServerStartUp(identifier) {
		return fmt.Errorf("failed to start agent server: %s", identifier)
	}
	if agentServerState.Load() != uint32(agentServerDetached) {
		agentServerState.Store(uint32(agentServerRunningAttached))
	}
	return nil
}

// AgentServerShutDown shuts down the MAA Agent Server.
// After AgentServerDetach, the native call cannot wait for the service thread
// and may race with its socket use. It does not make Release safe.
func AgentServerShutDown() {
	native.MaaAgentServerShutDown()
	if agentServerState.Load() != uint32(agentServerDetached) {
		agentServerState.Store(uint32(agentServerStopped))
	}
}

// AgentServerJoin waits for an attached agent service thread to end.
// After AgentServerDetach, the native call returns without waiting.
// AgentServerShutDown must still be called before Release in the attached case.
func AgentServerJoin() {
	native.MaaAgentServerJoin()
	agentServerState.CompareAndSwap(uint32(agentServerRunningAttached), uint32(agentServerJoined))
}

// AgentServerDetach detaches the service thread to run independently.
// It allows the service to run in the background without blocking.
// The current native API cannot confirm when a detached thread has exited,
// so Release remains blocked after this call.
func AgentServerDetach() {
	agentServerState.CompareAndSwap(uint32(agentServerRunningAttached), uint32(agentServerDetached))
	native.MaaAgentServerDetach()
}
