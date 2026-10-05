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
