package maa

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// TestAgentServer_ConcurrentShutDownCallsNativeOnce pins the shutdown
// serialization: when several goroutines request shutdown of a running server
// at the same time, the native shutdown function runs exactly once and the
// server settles in the terminal Closed state. Without the lifecycle mutex,
// concurrent callers can race past the Closed check and call native shutdown
// more than once.
func TestAgentServer_ConcurrentShutDownCallsNativeOnce(t *testing.T) {
	originalPhase := agentServerState.Load()
	agentServerState.Store(uint32(agentServerStopped))
	t.Cleanup(func() { agentServerState.Store(originalPhase) })

	var startupCalls, shutdownCalls atomic.Int64
	replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool {
		startupCalls.Add(1)
		return true
	})
	replaceNativeForTest(t, &native.MaaAgentServerShutDown, func() {
		shutdownCalls.Add(1)
	})

	require.NoError(t, AgentServerStartUp("concurrent-shutdown"))
	require.EqualValues(t, agentServerRunningAttached, agentServerState.Load())

	const goroutines = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			AgentServerShutDown()
		}()
	}
	close(start)
	wg.Wait()

	require.EqualValues(t, 1, shutdownCalls.Load(), "concurrent shutdown must call native ShutDown exactly once")
	require.EqualValues(t, agentServerClosed, agentServerState.Load())

	// The Closed state must be observable: startup is rejected before the
	// native path, so the startup call count must not change.
	require.ErrorIs(t, AgentServerStartUp("restart"), ErrClosed)
	require.EqualValues(t, 1, startupCalls.Load(), "closed server must not enter the native startup path")
}

// TestAgentServer_StartUpGuardAgainstRacingShutDown pins the startup guard:
// a ShutDown that slips in between StartUp's configuration check and its
// acquisition of the lifecycle mutex closes the singleton first, and StartUp
// must then stay out of the native startup path and return ErrClosed instead
// of starting on the closed singleton and overwriting the terminal state.
// The invariant is asserted through sequence stamps: whenever both native
// calls run in an iteration, startup must have run before shutdown.
func TestAgentServer_StartUpGuardAgainstRacingShutDown(t *testing.T) {
	originalPhase := agentServerState.Load()
	agentServerState.Store(uint32(agentServerStopped))
	t.Cleanup(func() { agentServerState.Store(originalPhase) })

	var seq, startupSeq, shutdownSeq atomic.Int64
	replaceNativeForTest(t, &native.MaaAgentServerStartUp, func(string) bool {
		startupSeq.Store(seq.Add(1))
		return true
	})
	replaceNativeForTest(t, &native.MaaAgentServerShutDown, func() {
		shutdownSeq.Store(seq.Add(1))
	})

	const iterations = 200
	for range iterations {
		prevStartup := startupSeq.Load()
		prevShutdown := shutdownSeq.Load()
		agentServerState.Store(uint32(agentServerStopped))

		var startErr error
		var wg sync.WaitGroup
		start := make(chan struct{})
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			startErr = AgentServerStartUp("startup-guard")
		}()
		go func() {
			defer wg.Done()
			<-start
			AgentServerShutDown()
		}()
		close(start)
		wg.Wait()

		curStartup := startupSeq.Load()
		curShutdown := shutdownSeq.Load()
		require.Greater(t, curShutdown, prevShutdown, "shutdown must run in every iteration")
		require.EqualValues(t, agentServerClosed, agentServerState.Load(), "every iteration must end Closed")
		if curStartup > prevStartup {
			require.Less(t, curStartup, curShutdown,
				"native startup ran after native shutdown had already closed the singleton")
			require.NoError(t, startErr)
		} else {
			require.ErrorIs(t, startErr, ErrClosed,
				"startup that lost the race to shutdown must report ErrClosed")
		}
	}
}
