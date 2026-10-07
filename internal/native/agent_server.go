package native

import (
	"fmt"
	"runtime"
)

var maaAgentServer uintptr

const maaAgentServerName = "MaaAgentServer"

// Agent server functions. The server runs in its own process and serves
// custom recognition and action implementations to the host that connects
// through an agent client. MaaAgentServerRegisterCustomRecognition and
// MaaAgentServerRegisterCustomAction register those implementations under
// the pipeline custom_recognition and custom_action field names; the native
// layer rejects duplicate names. The Add*Sink functions receive the events
// the connected client forwards for its resource, controller, tasker or
// context through a MaaEventCallback and return a sink ID, 0 on failure.
//
// MaaAgentServerStartUp starts the service in a separate native thread and
// returns after starting it; the identifier is the rendezvous name matched
// with the agent client, an IPC socket named after the identifier or, for a
// purely numeric identifier in 1-65535, the TCP port the client binds. An
// empty identifier fails to start. MaaAgentServerShutDown requests a stop
// and, while the service thread is attached, waits for it to exit before
// closing the sockets; MaaAgentServerJoin waits for that thread to end
// without requesting a stop; MaaAgentServerDetach abandons the thread,
// giving up the ability to wait for or observe its exit.
var (
	MaaAgentServerRegisterCustomRecognition func(name string, recognition MaaCustomRecognitionCallback, transArg uintptr) bool
	MaaAgentServerRegisterCustomAction      func(name string, action MaaCustomActionCallback, transArg uintptr) bool
	MaaAgentServerAddResourceSink           func(sink MaaEventCallback, transArg uintptr) int64
	MaaAgentServerAddControllerSink         func(sink MaaEventCallback, transArg uintptr) int64
	MaaAgentServerAddTaskerSink             func(sink MaaEventCallback, transArg uintptr) int64
	MaaAgentServerAddContextSink            func(sink MaaEventCallback, transArg uintptr) int64
	MaaAgentServerStartUp                   func(identifier string) bool
	MaaAgentServerShutDown                  func()
	MaaAgentServerJoin                      func()
	MaaAgentServerDetach                    func()
)

var agentServerEntries = []Entry{
	{&MaaAgentServerRegisterCustomRecognition, "MaaAgentServerRegisterCustomRecognition"},
	{&MaaAgentServerRegisterCustomAction, "MaaAgentServerRegisterCustomAction"},
	{&MaaAgentServerAddResourceSink, "MaaAgentServerAddResourceSink"},
	{&MaaAgentServerAddControllerSink, "MaaAgentServerAddControllerSink"},
	{&MaaAgentServerAddTaskerSink, "MaaAgentServerAddTaskerSink"},
	{&MaaAgentServerAddContextSink, "MaaAgentServerAddContextSink"},
	{&MaaAgentServerStartUp, "MaaAgentServerStartUp"},
	{&MaaAgentServerShutDown, "MaaAgentServerShutDown"},
	{&MaaAgentServerJoin, "MaaAgentServerJoin"},
	{&MaaAgentServerDetach, "MaaAgentServerDetach"},
}

func getMaaAgentServerLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaAgentServer.dylib"
	case "linux", "android":
		return "libMaaAgentServer.so"
	case "windows":
		return "MaaAgentServer.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}
