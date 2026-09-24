package native

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var maaAgentServer uintptr

const maaAgentServerName = "MaaAgentServer"

// MaaShutdownCallback 对应 C 侧的 typedef：
//   typedef void(MAA_CALL* MaaShutdownCallback)(void* trans_arg);
//
// AgentServer 收到远端 ShutDownRequest 时、停止消息循环前调用。
// 回调返回后才发送 ShutDownResponse，宿主可阻塞等待收尾完成。
//
// 注意：C 侧返回 void，但 Go 声明返回 uintptr 并在 trampoline 中返回 0——
// 这是 purego 在 Windows 上的硬性要求（syscall.NewCallback 强制回调
// 必须有一个 uintptr 大小的返回值，否则 panic）。C 侧会忽略这个返回值。
// 仓库内 MaaEventCallback 也是同样的 workaround（见 framework.go:20）。
//
// Corresponds to the C typedef:
//   typedef void(MAA_CALL* MaaShutdownCallback)(void* trans_arg);
//
// Invoked by AgentServer when a remote ShutDownRequest arrives, before the
// message loop stops. The ShutDownResponse is sent only after this callback
// returns, so the host may block inside to finish cleanup.
//
// Note: the C side returns void, but Go declares a uintptr return that the
// trampoline sets to 0 — purego on Windows requires exactly one uintptr-sized
// return value (syscall.NewCallback panics otherwise). C ignores the value.
// MaaEventCallback uses the same workaround (see framework.go:20).
type MaaShutdownCallback func(transArg uintptr) uintptr

var (
	MaaAgentServerRegisterCustomRecognition func(name string, recognition MaaCustomRecognitionCallback, transArg unsafe.Pointer) bool
	MaaAgentServerRegisterCustomAction      func(name string, action MaaCustomActionCallback, transArg unsafe.Pointer) bool
	MaaAgentServerAddResourceSink           func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddControllerSink         func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddTaskerSink             func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerAddContextSink            func(sink MaaEventCallback, transArg unsafe.Pointer) int64
	MaaAgentServerSetShutdownCallback       func(callback MaaShutdownCallback, transArg unsafe.Pointer) bool
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
	{&MaaAgentServerSetShutdownCallback, "MaaAgentServerSetShutdownCallback"},
	{&MaaAgentServerStartUp, "MaaAgentServerStartUp"},
	{&MaaAgentServerShutDown, "MaaAgentServerShutDown"},
	{&MaaAgentServerJoin, "MaaAgentServerJoin"},
	{&MaaAgentServerDetach, "MaaAgentServerDetach"},
}

func initAgentServer(libDir string) error {
	libName := getMaaAgentServerLibrary()
	libPath := filepath.Join(libDir, libName)

	handle, err := openLibrary(libPath)
	if err != nil {
		return &LibraryLoadError{
			LibraryName: maaAgentServerName,
			LibraryPath: libPath,
			Err:         err,
		}
	}

	maaAgentServer = handle

	registerAgentServer()

	return nil
}

func getMaaAgentServerLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaAgentServer.dylib"
	case "linux":
		return "libMaaAgentServer.so"
	case "windows":
		return "MaaAgentServer.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func registerAgentServer() {
	for _, entry := range agentServerEntries {
		purego.RegisterLibFunc(entry.ptrToFunc, maaAgentServer, entry.name)
	}
}

func releaseAgentServer() error {
	err := unloadLibrary(maaAgentServer)
	if err != nil {
		return err
	}

	unregisterAgentServer()

	return nil
}

func unregisterAgentServer() {
	for _, entry := range agentServerEntries {
		clearFuncVar(entry.ptrToFunc)
	}
}
