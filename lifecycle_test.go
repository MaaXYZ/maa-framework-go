//go:build !race

package maa

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// lifecycleHelperEnv marks a child invocation of the root test binary that runs
// the lifecycle scenario in isolation. Init and Release mutate package-level
// state, so exercising their transitions in-process would corrupt the shared
// root test runtime; a dedicated child process keeps sibling tests unaffected.
const lifecycleHelperEnv = "MAA_LIFECYCLE_TEST_HELPER"

// lifecycleDetachHelperEnv marks a child invocation that runs the detached
// Agent Server scenario. Detach leaves a terminal state that Release can never
// clear, so the scenario needs its own process instead of joining the shared
// lifecycle helper.
const lifecycleDetachHelperEnv = "MAA_LIFECYCLE_DETACH_TEST_HELPER"

// TestLifecycleInitReleaseBoundaries verifies the Init/Release contract from an
// already-initialized process. TestMain initializes the framework in the child
// before the scenario body starts, and every transition runs sequentially.
func TestLifecycleInitReleaseBoundaries(t *testing.T) {
	if os.Getenv(lifecycleHelperEnv) == "1" {
		runLifecycleHelper(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestLifecycleInitReleaseBoundaries$", "-test.count=1")
	cmd.Env = append(os.Environ(), lifecycleHelperEnv+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("lifecycle helper process failed: %v\n%s", err, output)
	}
}

// TestLifecycleDetachedAgentServerKeepsReleaseBlocked verifies that Release
// stays blocked after a detached Agent Server is shut down: the native API
// cannot confirm when the detached thread has exited, so the guard is terminal
// for the process. Native start/detach/shutdown calls are stubbed because a
// real detached ShutDown may race with ZMQ cleanup.
func TestLifecycleDetachedAgentServerKeepsReleaseBlocked(t *testing.T) {
	if os.Getenv(lifecycleDetachHelperEnv) == "1" {
		runLifecycleDetachHelper(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestLifecycleDetachedAgentServerKeepsReleaseBlocked$", "-test.count=1")
	cmd.Env = append(os.Environ(), lifecycleDetachHelperEnv+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("lifecycle detach helper process failed: %v\n%s", err, output)
	}
}

// runLifecycleHelper drives the transitions in order inside the child process.
// It relies on TestMain having already initialized the framework.
func runLifecycleHelper(t *testing.T) {
	t.Helper()

	// TestMain initialized the framework before this test ran.
	require.True(t, IsInited(), "TestMain did not leave the framework initialized")
	require.Zero(t, liveNativeObjects.Load(), "TestMain leaked native objects")

	// Repeating Init in an initialized process is harmless.
	require.NoError(t, Init())
	require.True(t, IsInited())

	// A live owned native object blocks Release without unloading anything.
	tasker, err := NewTasker()
	require.NoError(t, err)
	require.EqualValues(t, 1, liveNativeObjects.Load())
	require.ErrorIs(t, Release(), ErrLibraryInUse)
	require.True(t, IsInited(), "failed Release must leave the framework initialized")

	// Destroying the object lets the libraries unload.
	require.NoError(t, tasker.Destroy())
	require.Zero(t, liveNativeObjects.Load())
	require.NoError(t, Release())
	require.False(t, IsInited())

	// Release is idempotent with nothing loaded.
	require.NoError(t, Release())
	require.False(t, IsInited())

	// A missing library directory fails as a LibraryLoadError and loads nothing.
	missingDir := filepath.Join(t.TempDir(), "missing")
	err = Init(WithLibDir(missingDir))
	var loadErr *LibraryLoadError
	require.ErrorAs(t, err, &loadErr)
	require.False(t, IsInited(), "failed Init must not leave an initialized state")
	require.Zero(t, liveNativeObjects.Load())

	// A post-load option failure rolls the loaded libraries back.
	err = Init(WithLogDir(""))
	require.ErrorIs(t, err, ErrEmptyLogDir)
	require.False(t, IsInited(), "rolled-back Init must not leave an initialized state")
	require.Zero(t, liveNativeObjects.Load())

	// A later Init after failures succeeds.
	require.NoError(t, Init(WithLogDir(t.TempDir()), WithStdoutLevel(LoggingLevelOff)))
	require.True(t, IsInited())

	// A running Agent Server also blocks Release until it is shut down. The
	// native start/join/shutdown calls are stubbed so the child stays isolated.
	oldStartUp, oldJoin, oldShutDown := native.MaaAgentServerStartUp, native.MaaAgentServerJoin, native.MaaAgentServerShutDown
	joinCalls := 0
	native.MaaAgentServerStartUp = func(string) bool { return true }
	native.MaaAgentServerJoin = func() { joinCalls++ }
	native.MaaAgentServerShutDown = func() {}

	require.NoError(t, AgentServerStartUp("lifecycle-test"))
	require.ErrorIs(t, Release(), ErrLibraryInUse)
	require.True(t, IsInited(), "failed Release must leave the framework initialized")

	// A normal Join only confirms the service thread ended; the native
	// shutdown sequence is still pending, so Release stays blocked.
	AgentServerJoin()
	require.Equal(t, 1, joinCalls)
	require.ErrorIs(t, Release(), ErrLibraryInUse)
	require.True(t, IsInited(), "failed Release must leave the framework initialized")

	// An explicit ShutDown is required before Release can unload.
	AgentServerShutDown()
	// Restore the real symbols while the libraries are still loaded.
	native.MaaAgentServerStartUp = oldStartUp
	native.MaaAgentServerJoin = oldJoin
	native.MaaAgentServerShutDown = oldShutDown

	require.NoError(t, Release())
	require.False(t, IsInited())

	// A shutdown failure must not leave IsInited reporting that the native
	// functions are usable: Shutdown may already have unloaded some libraries.
	require.NoError(t, Init(WithStdoutLevel(LoggingLevelOff)))
	require.True(t, IsInited())
	shutdownErr := errors.New("injected shutdown failure")
	actualShutdown := shutdownNativeLibraries
	shutdownNativeLibraries = func() error { return shutdownErr }
	require.ErrorIs(t, Release(), shutdownErr)
	require.False(t, IsInited())

	// Release still retries cleanup even when the package is no longer marked
	// initialized. The injected failure left the real libraries open.
	shutdownNativeLibraries = actualShutdown
	require.NoError(t, Release())
	require.False(t, IsInited())
	require.Zero(t, liveNativeObjects.Load())
}

// runLifecycleDetachHelper drives StartUp -> Detach -> ShutDown -> Release in
// the child process. TestMain has already initialized the framework. Detach is
// terminal for the Release guard, so nothing after the failed Release runs in
// this process.
func runLifecycleDetachHelper(t *testing.T) {
	t.Helper()

	require.True(t, IsInited(), "TestMain did not leave the framework initialized")
	require.Zero(t, liveNativeObjects.Load(), "TestMain leaked native objects")

	// Stub the native symbols: a real detached ShutDown must never run here
	// because the native API cannot join the detached thread.
	oldStartUp, oldDetach, oldShutDown := native.MaaAgentServerStartUp, native.MaaAgentServerDetach, native.MaaAgentServerShutDown
	native.MaaAgentServerStartUp = func(string) bool { return true }
	native.MaaAgentServerDetach = func() {}
	native.MaaAgentServerShutDown = func() {}
	defer func() {
		native.MaaAgentServerStartUp = oldStartUp
		native.MaaAgentServerDetach = oldDetach
		native.MaaAgentServerShutDown = oldShutDown
	}()

	// Intercept library unloading so a broken guard cannot dlclose code that a
	// detached native thread may still be executing.
	unloadAttempted := false
	actualShutdown := shutdownNativeLibraries
	shutdownNativeLibraries = func() error {
		unloadAttempted = true
		return nil
	}
	defer func() { shutdownNativeLibraries = actualShutdown }()

	require.NoError(t, AgentServerStartUp("lifecycle-detach-test"))
	AgentServerDetach()
	AgentServerShutDown()

	require.ErrorIs(t, Release(), ErrLibraryInUse,
		"Release must stay blocked while the Agent Server state is detached")
	require.False(t, unloadAttempted,
		"shutdownNativeLibraries must not run while the Agent Server is detached")
	require.True(t, IsInited(), "failed Release must leave the framework initialized")
	require.Zero(t, liveNativeObjects.Load())
}
