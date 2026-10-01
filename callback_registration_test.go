package maa

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/store"
	"github.com/stretchr/testify/require"
)

func replaceNativeForTest[T any](t *testing.T, target *T, replacement T) {
	t.Helper()
	original := *target
	*target = replacement
	t.Cleanup(func() { *target = original })
}

func eventCallbackCount() int {
	eventCallbacksMutex.RLock()
	defer eventCallbacksMutex.RUnlock()
	return len(eventCallbacks)
}

func TestCustomController_CreateFailureRollsBack(t *testing.T) {
	customControllerCallbacksAgentsMutex.RLock()
	before := len(customControllerCallbacksAgents)
	customControllerCallbacksAgentsMutex.RUnlock()
	replaceNativeForTest(t, &native.MaaCustomControllerCreate, func(unsafe.Pointer, uintptr) uintptr { return 0 })
	for range 2 {
		ctrl, err := NewBlankController()
		require.Error(t, err)
		require.Nil(t, ctrl)
	}
	customControllerCallbacksAgentsMutex.RLock()
	after := len(customControllerCallbacksAgents)
	customControllerCallbacksAgentsMutex.RUnlock()
	require.Equal(t, before, after)
}

func TestInstanceSinks_FailureRollsBack(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ctrl.Destroy()) })
	tasker, err := NewTasker()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, tasker.Destroy()) })
	for _, tc := range []struct {
		name   string
		target *func(uintptr, native.MaaEventCallback, uintptr) int64
		add    func() int64
	}{
		{"Resource", &native.MaaResourceAddSink, func() int64 { return res.AddSink(nil) }},
		{"Controller", &native.MaaControllerAddSink, func() int64 { return ctrl.AddSink(nil) }},
		{"Tasker", &native.MaaTaskerAddSink, func() int64 { return tasker.AddSink(nil) }},
		{"Context", &native.MaaTaskerAddContextSink, func() int64 { return tasker.AddContextSink(nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := eventCallbackCount()
			replaceNativeForTest(t, tc.target, func(uintptr, native.MaaEventCallback, uintptr) int64 { return 0 })
			for range 2 {
				require.Zero(t, tc.add())
			}
			require.Equal(t, before, eventCallbackCount())
		})
	}
	store.ResStore.Update(res.handle, func(v *store.ResStoreValue) { require.Empty(t, v.SinkIDToEventCallbackID) })
	store.CtrlStore.Update(ctrl.handle, func(v *store.CtrlStoreValue) { require.Empty(t, v.SinkIDToEventCallbackID) })
	store.TaskerStore.Update(tasker.handle, func(v *store.TaskerStoreValue) {
		require.Empty(t, v.SinkIDToEventCallbackID)
		require.Empty(t, v.ContextSinkIDToEventCallbackID)
	})
}

func TestResource_SinkTransactions(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })
	before := eventCallbackCount()
	entered, release := make(chan struct{}), make(chan struct{})
	unblock := jobConcurrencyRelease(release)
	defer unblock()
	var mu sync.Mutex
	registered := make(map[int64]uintptr)
	replaceNativeForTest(t, &native.MaaResourceAddSink, func(_ uintptr, _ native.MaaEventCallback, id uintptr) int64 {
		mu.Lock()
		registered[42] = id
		mu.Unlock()
		close(entered)
		<-release
		return 42
	})
	replaceNativeForTest(t, &native.MaaResourceClearSinks, func(uintptr) {
		mu.Lock()
		clear(registered)
		mu.Unlock()
	})
	replaceNativeForTest(t, &native.MaaResourceRemoveSink, func(_ uintptr, id int64) {
		mu.Lock()
		delete(registered, id)
		mu.Unlock()
	})
	addDone := make(chan int64, 1)
	go func() { addDone <- res.AddSink(nil) }()
	jobConcurrencyAwait(t, entered)
	clearStarted, clearDone := make(chan struct{}), make(chan struct{})
	go func() { close(clearStarted); res.ClearSinks(); close(clearDone) }()
	jobConcurrencyAwait(t, clearStarted)
	unblock()
	require.EqualValues(t, 42, jobConcurrencyAwait(t, addDone))
	jobConcurrencyAwait(t, clearDone)
	res.RemoveSink(42)
	res.RemoveSink(42)
	mu.Lock()
	remaining := len(registered)
	mu.Unlock()
	require.Zero(t, remaining)
	store.ResStore.Update(res.handle, func(v *store.ResStoreValue) { require.Empty(t, v.SinkIDToEventCallbackID) })
	require.Equal(t, before, eventCallbackCount())
}

func TestResource_CustomRegistrationTransactions(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })
	for _, tc := range []struct {
		name       string
		register   func() error
		unregister func() error
		clear      func() error
		exists     func(uint64) bool
		stored     func() uint64
		stub       func(*testing.T, func(uintptr) bool, func() bool)
	}{
		{"Action", func() error { return res.RegisterCustomAction("test", nil) },
			func() error { return res.UnregisterCustomAction("test") }, res.ClearCustomAction,
			func(id uint64) bool {
				customActionRunnerCallbackAgentsMutex.RLock()
				defer customActionRunnerCallbackAgentsMutex.RUnlock()
				_, ok := customActionRunnerCallbackAgents[id]
				return ok
			},
			func() (id uint64) {
				store.ResStore.Update(res.handle, func(v *store.ResStoreValue) { id = v.CustomActionsCallbackID["test"] })
				return
			},
			func(t *testing.T, reg func(uintptr) bool, accept func() bool) {
				replaceNativeForTest(t, &native.MaaResourceRegisterCustomAction, func(_ uintptr, _ string, _ native.MaaCustomActionCallback, id uintptr) bool { return reg(id) })
				replaceNativeForTest(t, &native.MaaResourceUnregisterCustomAction, func(uintptr, string) bool { return accept() })
				replaceNativeForTest(t, &native.MaaResourceClearCustomAction, func(uintptr) bool { return accept() })
			}},
		{"Recognition", func() error { return res.RegisterCustomRecognition("test", nil) },
			func() error { return res.UnregisterCustomRecognition("test") }, res.ClearCustomRecognition,
			func(id uint64) bool {
				customRecognitionRunnerCallbackAgentsMutex.RLock()
				defer customRecognitionRunnerCallbackAgentsMutex.RUnlock()
				_, ok := customRecognitionRunnerCallbackAgents[id]
				return ok
			},
			func() (id uint64) {
				store.ResStore.Update(res.handle, func(v *store.ResStoreValue) { id = v.CustomRecognizersCallbackID["test"] })
				return
			},
			func(t *testing.T, reg func(uintptr) bool, accept func() bool) {
				replaceNativeForTest(t, &native.MaaResourceRegisterCustomRecognition, func(_ uintptr, _ string, _ native.MaaCustomRecognitionCallback, id uintptr) bool { return reg(id) })
				replaceNativeForTest(t, &native.MaaResourceUnregisterCustomRecognition, func(uintptr, string) bool { return accept() })
				replaceNativeForTest(t, &native.MaaResourceClearCustomRecognition, func(uintptr) bool { return accept() })
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accept := true
			var attempted uint64
			tc.stub(t, func(id uintptr) bool { attempted = uint64(id); return accept }, func() bool { return accept })
			require.NoError(t, tc.register())
			first := tc.stored()
			accept = false
			require.Error(t, tc.register())
			require.False(t, tc.exists(attempted))
			require.Error(t, tc.unregister())
			require.Error(t, tc.clear())
			require.Equal(t, first, tc.stored())
			require.True(t, tc.exists(first))
			accept = true
			require.NoError(t, tc.register())
			require.False(t, tc.exists(first))
			require.Equal(t, attempted, tc.stored())
			require.NoError(t, tc.unregister())
			require.Error(t, tc.unregister())
			require.False(t, tc.exists(attempted))
			require.NoError(t, tc.register())
			require.NoError(t, tc.clear())
			require.NoError(t, tc.clear())
			require.Zero(t, tc.stored())
			require.False(t, tc.exists(attempted))
		})
	}
}

func TestAgentServer_ConfigurationBoundary(t *testing.T) {
	originalPhase := agentServerState.Load()
	agentServerState.Store(uint32(agentServerStopped))
	t.Cleanup(func() { agentServerState.Store(originalPhase) })
	for _, tc := range []struct {
		name   string
		target *func(native.MaaEventCallback, uintptr) int64
		add    func() int64
	}{
		{"Resource", &native.MaaAgentServerAddResourceSink, func() int64 { return AgentServerAddResourceSink(nil) }},
		{"Controller", &native.MaaAgentServerAddControllerSink, func() int64 { return AgentServerAddControllerSink(nil) }},
		{"Tasker", &native.MaaAgentServerAddTaskerSink, func() int64 { return AgentServerAddTaskerSink(nil) }},
		{"Context", &native.MaaAgentServerAddContextSink, func() int64 { return AgentServerAddContextSink(nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := eventCallbackCount()
			calls := 0
			replaceNativeForTest(t, tc.target, func(native.MaaEventCallback, uintptr) int64 { calls++; return 0 })
			agentServerState.Store(uint32(agentServerStopped))
			for range 2 {
				require.Zero(t, tc.add())
			}
			require.Equal(t, before, eventCallbackCount())
			agentServerState.Store(uint32(agentServerRunningAttached))
			require.Zero(t, tc.add())
			require.Equal(t, 2, calls)
		})
	}
	require.ErrorIs(t, AgentServerRegisterCustomAction("busy", nil), ErrInUse)
	require.ErrorIs(t, AgentServerRegisterCustomRecognition("busy", nil), ErrInUse)
	replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool { return false })
	agentServerState.Store(uint32(agentServerStopped))
	require.Error(t, AgentServerStartUp("failure"))
	require.EqualValues(t, agentServerStopped, agentServerState.Load())
	replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool { return true })
	require.NoError(t, AgentServerStartUp("success"))
	require.ErrorIs(t, AgentServerStartUp("again"), ErrInUse)
}

func TestAgentServer_ShutDownIsTerminal(t *testing.T) {
	originalPhase := agentServerState.Load()
	t.Cleanup(func() { agentServerState.Store(originalPhase) })
	for _, tc := range []struct {
		name  string
		phase agentServerPhase
	}{
		{"BeforeStartUp", agentServerStopped},
		{"RunningAttached", agentServerRunningAttached},
		{"Joined", agentServerJoined},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agentServerState.Store(uint32(tc.phase))
			shutdownCalls, startupCalls, configurationCalls := 0, 0, 0
			replaceNativeForTest(t, &native.MaaAgentServerShutDown, func() { shutdownCalls++ })
			replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool { startupCalls++; return true })
			replaceNativeForTest(t, &native.MaaAgentServerRegisterCustomAction, func(string, native.MaaCustomActionCallback, uintptr) bool {
				configurationCalls++
				return false
			})
			replaceNativeForTest(t, &native.MaaAgentServerRegisterCustomRecognition, func(string, native.MaaCustomRecognitionCallback, uintptr) bool {
				configurationCalls++
				return false
			})
			for _, target := range []*func(native.MaaEventCallback, uintptr) int64{
				&native.MaaAgentServerAddResourceSink,
				&native.MaaAgentServerAddControllerSink,
				&native.MaaAgentServerAddTaskerSink,
				&native.MaaAgentServerAddContextSink,
			} {
				replaceNativeForTest(t, target, func(native.MaaEventCallback, uintptr) int64 {
					configurationCalls++
					return 0
				})
			}

			AgentServerShutDown()
			require.EqualValues(t, agentServerClosed, agentServerState.Load())
			require.ErrorIs(t, AgentServerStartUp("restart"), ErrClosed)
			require.Zero(t, startupCalls, "must not enter the native startup path after shutdown")
			require.ErrorIs(t, AgentServerRegisterCustomAction("closed", nil), ErrClosed)
			require.ErrorIs(t, AgentServerRegisterCustomRecognition("closed", nil), ErrClosed)
			before := eventCallbackCount()
			require.Zero(t, AgentServerAddResourceSink(nil))
			require.Zero(t, AgentServerAddControllerSink(nil))
			require.Zero(t, AgentServerAddTaskerSink(nil))
			require.Zero(t, AgentServerAddContextSink(nil))
			require.Equal(t, before, eventCallbackCount())
			require.Zero(t, configurationCalls, "closed configuration must not call native functions")
			AgentServerShutDown()
			require.Equal(t, 1, shutdownCalls, "repeated shutdown must not call native functions")
		})
	}
}

func TestEventCallback_ClosedOwnerRejectsLateDispatch(t *testing.T) {
	state := newExternalHandleState(789)
	calls := 0
	id := registerEventCallback(resourceSinkFunc(func(*Resource) { calls++ }), state)
	defer unregisterEventCallback(id)
	require.NoError(t, state.close())
	message := append([]byte(EventResourceLoading.Starting()), 0)
	details := []byte("{}\x00")
	_MaaEventCallbackAgent(789, &message[0], &details[0], uintptr(id))
	require.Zero(t, calls)
}

func TestAgentServer_CustomReplacement(t *testing.T) {
	phase := agentServerState.Load()
	agentServerState.Store(uint32(agentServerStopped))
	defer agentServerState.Store(phase)
	for _, tc := range []struct {
		name     string
		register func(string) error
		ids      map[string]uint64
		exists   func(uint64) bool
		remove   func(uint64)
		stub     func(*testing.T, func(uintptr) bool)
	}{
		{"Action", func(name string) error { return AgentServerRegisterCustomAction(name, nil) }, agentServerActionIDs,
			func(id uint64) bool {
				customActionRunnerCallbackAgentsMutex.RLock()
				defer customActionRunnerCallbackAgentsMutex.RUnlock()
				_, ok := customActionRunnerCallbackAgents[id]
				return ok
			},
			func(id uint64) { unregisterCustomAction(id) },
			func(t *testing.T, f func(uintptr) bool) {
				replaceNativeForTest(t, &native.MaaAgentServerRegisterCustomAction, func(_ string, _ native.MaaCustomActionCallback, id uintptr) bool { return f(id) })
			}},
		{"Recognition", func(name string) error { return AgentServerRegisterCustomRecognition(name, nil) }, agentServerRecognitionIDs,
			func(id uint64) bool {
				customRecognitionRunnerCallbackAgentsMutex.RLock()
				defer customRecognitionRunnerCallbackAgentsMutex.RUnlock()
				_, ok := customRecognitionRunnerCallbackAgents[id]
				return ok
			},
			func(id uint64) { unregisterCustomRecognition(id) },
			func(t *testing.T, f func(uintptr) bool) {
				replaceNativeForTest(t, &native.MaaAgentServerRegisterCustomRecognition, func(_ string, _ native.MaaCustomRecognitionCallback, id uintptr) bool { return f(id) })
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := t.Name()
			defer func() { tc.remove(tc.ids[name]); delete(tc.ids, name) }()
			accept := true
			var attempted uint64
			tc.stub(t, func(id uintptr) bool { attempted = uint64(id); return accept })
			require.NoError(t, tc.register(name))
			first := tc.ids[name]
			accept = false
			require.Error(t, tc.register(name))
			require.Equal(t, first, tc.ids[name])
			require.False(t, tc.exists(attempted))
			require.True(t, tc.exists(first))
			accept = true
			require.NoError(t, tc.register(name))
			require.Equal(t, attempted, tc.ids[name])
			require.False(t, tc.exists(first))
			require.True(t, tc.exists(attempted))
		})
	}
}

func TestAgentServer_ConfigurationBeforeStartUp(t *testing.T) {
	phase := agentServerState.Load()
	agentServerState.Store(uint32(agentServerStopped))
	defer agentServerState.Store(phase)
	name := t.Name()
	defer func() { unregisterCustomAction(agentServerActionIDs[name]); delete(agentServerActionIDs, name) }()
	entered, release := make(chan struct{}), make(chan struct{})
	unblock := jobConcurrencyRelease(release)
	defer unblock()
	replaceNativeForTest(t, &native.MaaAgentServerRegisterCustomAction, func(string, native.MaaCustomActionCallback, uintptr) bool {
		close(entered)
		<-release
		return true
	})
	committed := make(chan bool, 1)
	replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool {
		committed <- agentServerActionIDs[name] != 0
		return true
	})
	registered, started := make(chan error, 1), make(chan error, 1)
	go func() { registered <- AgentServerRegisterCustomAction(name, nil) }()
	jobConcurrencyAwait(t, entered)
	go func() { started <- AgentServerStartUp(name) }()
	unblock()
	require.NoError(t, jobConcurrencyAwait(t, registered))
	require.NoError(t, jobConcurrencyAwait(t, started))
	require.True(t, jobConcurrencyAwait(t, committed))
}
