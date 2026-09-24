package maa

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
)

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

// shutdownCallback 保存当前注册的关闭回调。
// C 侧只存 trampoline 函数指针（transArg 传 nil，不使用），
// 真正的 Go 回调函数保存在这个全局变量里，用 RWMutex 保护并发读写。
// （与其他回调的 map+ID 模式相比，关闭回调只有一个，不需要 map。）
//
// Holds the currently registered shutdown callback. The C side only stores
// a trampoline function pointer (transArg is nil, unused); the actual Go
// callback lives in this global variable, guarded by RWMutex.
// (Unlike other callbacks that use a map+ID pattern, there is only one
// shutdown callback, so a map is unnecessary.)
var (
	shutdownCallback      func()
	shutdownCallbackMutex sync.RWMutex
)

// _MaaShutdownCallbackAgent 是 C 侧调用的 trampoline。
//
// 当 AgentServer 收到 ShutDownRequest 时，C 层在消息循环线程里调用这个函数。
// 它返回后 C 层才会发送 ShutDownResponse，所以回调里可以安全地做耗时收尾
// （如发送在途通知）——客户端的 Disconnect() 会等到这里返回。
//
// 返回 0 是 purego/Windows 的硬性要求（回调必须有一个 uintptr 返回值），
// C 侧 void 返回会忽略它。
//
// Trampoline called from C on the server's message thread when a
// ShutDownRequest arrives. ShutDownResponse is not sent until this returns,
// so the callback can safely block for cleanup — the client's Disconnect()
// waits until this returns.
//
// Returning 0 satisfies purego/Windows (callbacks must have one uintptr
// return value); C ignores it since the typedef returns void.
func _MaaShutdownCallbackAgent(transArg uintptr) uintptr {
	shutdownCallbackMutex.RLock()
	callback := shutdownCallback
	shutdownCallbackMutex.RUnlock()

	if callback != nil {
		// recover 防止用户回调的 panic 穿透 C 边界（否则整进程 fatalpanic）。
		// recover 后 C 层照常发送 ShutDownResponse，优雅关停流程继续。
		//
		// Recover prevents a user-callback panic from crossing the C boundary
		// (which would fatally crash the process). After recovery, C still
		// sends the ShutDownResponse and graceful shutdown continues.
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("maa: panic in shutdown callback (recovered): %v\n", r)
			}
		}()
		callback()
	}
	return 0
}

// AgentServerSetShutdownCallback 注册关闭回调。
//
// AgentServer 收到 ShutDownRequest 时、停止消息循环前调用该回调；
// 回调返回后服务端才会回复 ShutDownResponse——所以回调里可以阻塞等待
// 收尾完成（如冲走在途通知），客户端的 Disconnect() 返回即代表收尾完毕。
//
// 回调在服务端消息循环线程中阻塞执行，须遵守以下约束：
//   - 不得 panic（已由 trampoline 内 recover 兜底，但仍应自行处理错误）
//   - 勿调用 AgentServerJoin 或 AgentServerShutDown（会等待自身线程导致死锁）
//   - 应尽快返回（客户端在同步等待，超时则先行放弃）
//
// 必须在 AgentServerStartUp 之前调用（运行中调用会返回错误）。
// 重复注册将覆盖之前的回调。
//
// Registers a shutdown callback. It is invoked when the AgentServer receives
// a ShutDownRequest, before the message loop stops; the ShutDownResponse is
// replied only after the callback returns — so the callback can block to
// finish cleanup (e.g. flushing in-flight notifications), and the client's
// Disconnect() returning means cleanup is complete.
//
// The callback runs blocking on the server's message thread and must follow
// these constraints:
//   - Must not panic (a recover in the trampoline acts as a safety net,
//     but errors should be handled properly)
//   - Do NOT call AgentServerJoin or AgentServerShutDown inside the callback
//     (they would wait for the calling thread itself, causing a deadlock)
//   - Should return as soon as possible (the client is synchronously
//     waiting and may time out)
//
// Must be called before AgentServerStartUp (returns an error while the
// server is running). Re-registration overwrites the previous callback.
func AgentServerSetShutdownCallback(callback func()) error {
	if callback == nil {
		return fmt.Errorf("callback must not be nil")
	}

	// 整个「写入→调C→失败恢复」放在同一个写锁临界区：
	// ① 保证注册失败时精确恢复旧值（而非置 nil 误杀已注册的回调）
	// ② 消灭 TOCTOU 窗口（写入新值与 C 接受之间的间隙，trampoline 不会读到中间态）
	// 该 native 调用不重入 Go、不 join 消息线程，持锁跨 FFI 安全。
	//
	// The entire "write → call C → roll back on failure" sequence runs in a
	// single write-lock critical section:
	// ① On failure, restores the previous value (instead of wiping to nil,
	//   which would silently destroy an already-registered callback)
	// ② Eliminates the TOCTOU window (the trampoline cannot observe an
	//   intermediate state between writing and C acceptance)
	// The native call does not re-enter Go or join the message thread,
	// so holding the lock across FFI is safe.
	shutdownCallbackMutex.Lock()
	prev := shutdownCallback
	shutdownCallback = callback

	ok := native.MaaAgentServerSetShutdownCallback(
		_MaaShutdownCallbackAgent,
		// transArg 传 nil——C 侧只存不解引用，真正的回调查找在 trampoline 里完成
		// transArg is nil — C stores it without dereferencing; the actual
		// callback lookup happens inside the trampoline
		nil,
	)
	if !ok {
		// 注册失败（如服务端已 start_up），恢复旧值
		// Registration failed (e.g. server already started), restore previous value
		shutdownCallback = prev
		shutdownCallbackMutex.Unlock()
		return fmt.Errorf("failed to set shutdown callback (must be called before AgentServerStartUp)")
	}
	shutdownCallbackMutex.Unlock()
	return nil
}

// AgentServerStartUp starts the MAA Agent Server with the given identifier.
// The identifier is used to match with AgentClient.
func AgentServerStartUp(identifier string) error {
	if !native.MaaAgentServerStartUp(identifier) {
		return fmt.Errorf("failed to start agent server: %s", identifier)
	}
	return nil
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
