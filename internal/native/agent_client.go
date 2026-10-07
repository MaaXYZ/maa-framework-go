package native

import (
	"fmt"
	"runtime"
)

var maaAgentClient uintptr

const maaAgentClientName = "MaaAgentClient"

// Agent client functions. The client binds the rendezvous endpoint that an
// agent server connects to: MaaAgentClientCreateV2 takes the identifier as
// a caller-created MaaStringBuffer handle (an IPC socket named after the
// identifier, where a purely numeric identifier in 1-65535 acts as a TCP
// port and an empty identifier is auto-generated), MaaAgentClientCreateTcp
// binds 127.0.0.1 on the given port instead, with port 0 letting the OS
// choose, and MaaAgentClientDestroy frees the client.
// MaaAgentClientIdentifier writes the effective identifier, the bound port
// string in TCP mode, into the caller-created MaaStringBuffer.
//
// MaaAgentClientBindResource binds the local resource that serves the
// connected server, and the Register*Sink functions forward the lifecycle
// events of the given local resource, controller or tasker to the server.
// MaaAgentClientConnect performs the protocol handshake once the server has
// connected and fails without a bound resource; Disconnect tears the
// connection down, Connected reports the connection state, and Alive probes
// the local communication channel, which is not a heartbeat.
//
// MaaAgentClientSetTimeout sets the timeout in milliseconds for each
// communication wait rather than for a whole operation. The Get*List
// functions fill the caller-created MaaStringListBuffer with the custom
// recognition and action names registered on the connected server; the
// lists are empty until Connect succeeds.
var (
	MaaAgentClientCreateV2                 func(identifier uintptr) uintptr
	MaaAgentClientCreateTcp                func(port uint16) uintptr
	MaaAgentClientDestroy                  func(client uintptr)
	MaaAgentClientIdentifier               func(client uintptr, identifier uintptr) bool
	MaaAgentClientBindResource             func(client uintptr, res uintptr) bool
	MaaAgentClientRegisterResourceSink     func(client uintptr, res uintptr) bool
	MaaAgentClientRegisterControllerSink   func(client uintptr, ctrl uintptr) bool
	MaaAgentClientRegisterTaskerSink       func(client uintptr, tasker uintptr) bool
	MaaAgentClientConnect                  func(client uintptr) bool
	MaaAgentClientDisconnect               func(client uintptr) bool
	MaaAgentClientConnected                func(client uintptr) bool
	MaaAgentClientAlive                    func(client uintptr) bool
	MaaAgentClientSetTimeout               func(client uintptr, milliseconds int64) bool
	MaaAgentClientGetCustomRecognitionList func(client uintptr, buffer uintptr) bool
	MaaAgentClientGetCustomActionList      func(client uintptr, buffer uintptr) bool
)

var agentClientEntries = []Entry{
	{&MaaAgentClientCreateV2, "MaaAgentClientCreateV2"},
	{&MaaAgentClientCreateTcp, "MaaAgentClientCreateTcp"},
	{&MaaAgentClientDestroy, "MaaAgentClientDestroy"},
	{&MaaAgentClientIdentifier, "MaaAgentClientIdentifier"},
	{&MaaAgentClientBindResource, "MaaAgentClientBindResource"},
	{&MaaAgentClientRegisterResourceSink, "MaaAgentClientRegisterResourceSink"},
	{&MaaAgentClientRegisterControllerSink, "MaaAgentClientRegisterControllerSink"},
	{&MaaAgentClientRegisterTaskerSink, "MaaAgentClientRegisterTaskerSink"},
	{&MaaAgentClientConnect, "MaaAgentClientConnect"},
	{&MaaAgentClientDisconnect, "MaaAgentClientDisconnect"},
	{&MaaAgentClientConnected, "MaaAgentClientConnected"},
	{&MaaAgentClientAlive, "MaaAgentClientAlive"},
	{&MaaAgentClientSetTimeout, "MaaAgentClientSetTimeout"},
	{&MaaAgentClientGetCustomRecognitionList, "MaaAgentClientGetCustomRecognitionList"},
	{&MaaAgentClientGetCustomActionList, "MaaAgentClientGetCustomActionList"},
}

func getMaaAgentClientLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaAgentClient.dylib"
	case "linux", "android":
		return "libMaaAgentClient.so"
	case "windows":
		return "MaaAgentClient.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}
