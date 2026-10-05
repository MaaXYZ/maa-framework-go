package maa

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

type agentServerPhase uint32

const (
	agentServerStopped agentServerPhase = iota
	agentServerRunningAttached
	agentServerJoined
	agentServerDetached
	agentServerClosed
)

// A detached native thread cannot be joined or observed through the current
// MaaAgentServer API, so agentServerDetached is a terminal state for Release.
// ShutDown closes the native singleton's context without a reset API, so
// agentServerClosed is terminal for startup and configuration, even across Init.
var agentServerState atomic.Uint32

var (
	agentServerConfigurationMu sync.Mutex
	agentServerRecognitionIDs  = make(map[string]uint64)
	agentServerActionIDs       = make(map[string]uint64)
)

// agentServerLifecycleMu serializes the server lifecycle operations (StartUp,
// ShutDown, Join, Detach) so concurrent callers cannot double-join the native
// service thread or race its state transitions. Only StartUp also holds
// agentServerConfigurationMu, always acquired first, so no lock cycle exists.
var agentServerLifecycleMu sync.Mutex

func lockAgentServerConfiguration() (func(), error) {
	agentServerConfigurationMu.Lock()
	if agentServerState.Load() == uint32(agentServerClosed) {
		agentServerConfigurationMu.Unlock()
		return nil, ErrClosed
	}
	if agentServerState.Load() != uint32(agentServerStopped) {
		agentServerConfigurationMu.Unlock()
		return nil, ErrInUse
	}
	return agentServerConfigurationMu.Unlock, nil
}

// AgentServerRegisterCustomRecognition registers a custom recognition runner.
// The name should match the custom_recognition field in Pipeline.
// Configure before StartUp. It returns ErrInUse while the server is active,
// joined, or detached, and ErrClosed after an attached server has shut down.
func AgentServerRegisterCustomRecognition(name string, recognition CustomRecognitionRunner) error {
	unlock, err := lockAgentServerConfiguration()
	if err != nil {
		return err
	}
	defer unlock()

	id := registerCustomRecognition(recognition)

	ok := native.MaaAgentServerRegisterCustomRecognition(
		name,
		_MaaCustomRecognitionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if !ok {
		unregisterCustomRecognition(id)
		return fmt.Errorf("failed to register custom recognition: %s", name)
	}
	if oldID := agentServerRecognitionIDs[name]; oldID != 0 {
		unregisterCustomRecognition(oldID)
	}
	agentServerRecognitionIDs[name] = id
	return nil
}

// AgentServerRegisterCustomAction registers a custom action runner.
// The name should match the custom_action field in Pipeline.
// Configure before StartUp. It returns ErrInUse while the server is active,
// joined, or detached, and ErrClosed after an attached server has shut down.
func AgentServerRegisterCustomAction(name string, action CustomActionRunner) error {
	unlock, err := lockAgentServerConfiguration()
	if err != nil {
		return err
	}
	defer unlock()

	id := registerCustomAction(action)

	ok := native.MaaAgentServerRegisterCustomAction(
		name,
		_MaaCustomActionCallbackAgent,
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		uintptr(id),
	)
	if !ok {
		unregisterCustomAction(id)
		return fmt.Errorf("failed to register custom action: %s", name)
	}
	if oldID := agentServerActionIDs[name]; oldID != 0 {
		unregisterCustomAction(oldID)
	}
	agentServerActionIDs[name] = id
	return nil
}

// AgentServerAddResourceSink adds a resource event callback sink and returns the sink ID.
// Configure before StartUp. It returns 0 on failure, including after ShutDown.
func AgentServerAddResourceSink(sink ResourceEventSink) int64 {
	return addAgentServerSink(sink, native.MaaAgentServerAddResourceSink)
}

// AgentServerAddControllerSink adds a controller event callback sink and returns the sink ID.
// Configure before StartUp. It returns 0 on failure, including after ShutDown.
func AgentServerAddControllerSink(sink ControllerEventSink) int64 {
	return addAgentServerSink(sink, native.MaaAgentServerAddControllerSink)
}

// AgentServerAddTaskerSink adds a tasker event callback sink and returns the sink ID.
// Configure before StartUp. It returns 0 on failure, including after ShutDown.
func AgentServerAddTaskerSink(sink TaskerEventSink) int64 {
	return addAgentServerSink(sink, native.MaaAgentServerAddTaskerSink)
}

// AgentServerAddContextSink adds a context event callback sink and returns the sink ID.
// Configure before StartUp. It returns 0 on failure, including after ShutDown.
func AgentServerAddContextSink(sink ContextEventSink) int64 {
	return addAgentServerSink(sink, native.MaaAgentServerAddContextSink)
}

func addAgentServerSink(sink any, add func(native.MaaEventCallback, uintptr) int64) int64 {
	unlock, err := lockAgentServerConfiguration()
	if err != nil {
		return 0
	}
	defer unlock()
	id := registerEventCallback(sink)
	sinkID := add(_MaaEventCallbackAgent, uintptr(id))
	if sinkID == 0 {
		unregisterEventCallback(id)
	}
	return sinkID
}

// AgentServerStartUp starts the MAA Agent Server in a separate native thread.
// It returns after starting the thread; call AgentServerJoin only when the
// caller needs to wait for the service to end. The identifier is used to match
// with AgentClient.
// Concurrent server lifecycle operations are serialized internally, but they
// must not be called from server callbacks. After an attached server's
// ShutDown, StartUp returns ErrClosed for the rest of the process, even after
// Release and Init, because the native singleton's communication context
// cannot be reset. It returns ErrInUse while the server is running, joined,
// or detached.
func AgentServerStartUp(identifier string) error {
	unlock, err := lockAgentServerConfiguration()
	if err != nil {
		return err
	}
	defer unlock()

	agentServerLifecycleMu.Lock()
	defer agentServerLifecycleMu.Unlock()

	if !native.MaaAgentServerStartUp(identifier) {
		return fmt.Errorf("failed to start agent server: %s", identifier)
	}
	if agentServerState.Load() != uint32(agentServerDetached) {
		agentServerState.Store(uint32(agentServerRunningAttached))
	}
	return nil
}

// AgentServerShutDown requests that the MAA Agent Server stop. If its service
// thread is still attached, the native call waits for the thread to exit before
// closing its sockets. After AgentServerDetach, it cannot wait for the thread
// and closing the sockets may race with their use. It does not make Release
// safe after Detach.
// Without prior Detach, ShutDown permanently prevents startup and
// configuration, even if called before StartUp; repeated calls are no-ops.
func AgentServerShutDown() {
	agentServerLifecycleMu.Lock()
	defer agentServerLifecycleMu.Unlock()

	if agentServerState.Load() == uint32(agentServerClosed) {
		return
	}
	native.MaaAgentServerShutDown()
	if agentServerState.Load() != uint32(agentServerDetached) {
		agentServerState.Store(uint32(agentServerClosed))
	}
}

// AgentServerJoin waits for an attached agent service thread to end. It does
// not request that the service stop, so it may block while the service runs.
// After AgentServerDetach, it returns without waiting. Even after Join returns
// for an attached thread, AgentServerShutDown must be called before Release.
func AgentServerJoin() {
	agentServerLifecycleMu.Lock()
	defer agentServerLifecycleMu.Unlock()

	native.MaaAgentServerJoin()
	agentServerState.CompareAndSwap(uint32(agentServerRunningAttached), uint32(agentServerJoined))
}

// AgentServerDetach detaches the service thread started by AgentServerStartUp.
// StartUp already runs the service in a separate thread; Detach gives up the
// ability to wait for that thread to exit. Use it only when the service is
// intended to last for the lifetime of the process and Release is not needed.
// After Detach, AgentServerJoin cannot wait for the thread, and
// AgentServerShutDown cannot confirm its exit. Release returns ErrLibraryInUse
// for the rest of the process, even after Join or ShutDown.
func AgentServerDetach() {
	agentServerLifecycleMu.Lock()
	defer agentServerLifecycleMu.Unlock()

	agentServerState.CompareAndSwap(uint32(agentServerRunningAttached), uint32(agentServerDetached))
	native.MaaAgentServerDetach()
}
