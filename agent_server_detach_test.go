//go:build !race

package maa

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Role markers for child invocations of this test binary. The parent process
// orchestrates one server child and one client child so the Release guard is
// exercised without any client/resource handle living in the server process.
const (
	agentServerDetachRoleEnv = "MAA_AGENT_SERVER_DETACH_TEST_ROLE"
	agentServerDetachIDEnv   = "MAA_AGENT_SERVER_DETACH_TEST_ID"
	agentServerDetachDirEnv  = "MAA_AGENT_SERVER_DETACH_TEST_DIR"

	agentServerDetachRoleServer = "server"
	agentServerDetachRoleClient = "client"
)

// Coordination files shared through the temp directory named by
// MAA_AGENT_SERVER_DETACH_TEST_DIR.
const (
	agentServerDetachServerReadyPath     = "server-ready"
	agentServerDetachClientConnectedPath = "client-connected"
	agentServerDetachReleaseDonePath     = "release-done"
	agentServerDetachReportPath          = "release-report.json"
	agentServerDetachServerLogPath       = "server.log"
	agentServerDetachClientLogPath       = "client.log"
)

const (
	agentServerDetachFileWait   = 10 * time.Second
	agentServerDetachProcWait   = 15 * time.Second
	agentServerDetachKillWait   = 5 * time.Second
	agentServerDetachPollPeriod = 20 * time.Millisecond
)

// agentServerDetachReport records the Release guard outcome in the server
// helper so the parent can assert even if helper cleanup is interrupted.
type agentServerDetachReport struct {
	LiveNativeObjects     int64  `json:"live_native_objects"`
	ReleaseError          string `json:"release_error"`
	ReleaseIsLibraryInUse bool   `json:"release_is_library_in_use"`
	UnloadAttempted       bool   `json:"unload_attempted"`
}

// TestAgentServerDetachJoinKeepsReleaseBlocked proves that a detached Agent
// Server which is still serving a client must keep Release blocked. The server
// child runs StartUp -> Detach -> Join -> Release and owns no native client or
// resource handles. A separate client child connects with the public Go
// wrappers and stays connected until after Release returns.
func TestAgentServerDetachJoinKeepsReleaseBlocked(t *testing.T) {
	switch os.Getenv(agentServerDetachRoleEnv) {
	case agentServerDetachRoleServer:
		runAgentServerDetachServerHelper(t)
		return
	case agentServerDetachRoleClient:
		runAgentServerDetachClientHelper(t)
		return
	}

	dir := t.TempDir()
	identifier, err := newAgentServerDetachIdentifier()
	require.NoError(t, err)

	serverLogPath := filepath.Join(dir, agentServerDetachServerLogPath)
	clientLogPath := filepath.Join(dir, agentServerDetachClientLogPath)
	server := startAgentServerDetachChild(t, agentServerDetachRoleServer, identifier, dir, serverLogPath)
	client := startAgentServerDetachChild(t, agentServerDetachRoleClient, identifier, dir, clientLogPath)

	serverErr := server.wait(t)
	clientErr := client.wait(t)
	if serverErr != nil {
		t.Fatalf("server helper failed: %v\n%s", serverErr, readAgentServerDetachLog(serverLogPath))
	}
	if clientErr != nil {
		t.Fatalf("client helper failed: %v\n%s", clientErr, readAgentServerDetachLog(clientLogPath))
	}

	report := readAgentServerDetachReport(t, filepath.Join(dir, agentServerDetachReportPath), serverLogPath)
	assertAgentServerDetachReleaseBlocked(t, report)
}

func assertAgentServerDetachReleaseBlocked(t *testing.T, report agentServerDetachReport) {
	t.Helper()

	require.Zero(t, report.LiveNativeObjects,
		"server subprocess must own no native handles at Release")

	var releaseErr error
	if report.ReleaseError != "" {
		if report.ReleaseIsLibraryInUse {
			releaseErr = ErrLibraryInUse
		} else {
			releaseErr = errors.New(report.ReleaseError)
		}
	}
	require.ErrorIs(t, releaseErr, ErrLibraryInUse,
		"Release after Detach and Join returned %v; unload attempted: %t",
		releaseErr, report.UnloadAttempted)
	require.False(t, report.UnloadAttempted,
		"shutdownNativeLibraries must not run while the Agent Server is still active")
}

// runAgentServerDetachServerHelper starts a real Agent Server, detaches it, and
// joins so the native service thread keeps running independently. It then waits
// until a separate client has finished its handshake and calls Release while
// that client is still connected. The server owns no native client or resource
// handles, so liveNativeObjects must be zero at the Release call.
func runAgentServerDetachServerHelper(t *testing.T) {
	t.Helper()

	require.True(t, IsInited())
	require.Zero(t, liveNativeObjects.Load())

	dir := os.Getenv(agentServerDetachDirEnv)
	identifier := os.Getenv(agentServerDetachIDEnv)
	require.NotEmpty(t, dir)
	require.NotEmpty(t, identifier)

	serverReadyPath := filepath.Join(dir, agentServerDetachServerReadyPath)
	clientConnectedPath := filepath.Join(dir, agentServerDetachClientConnectedPath)
	releaseDonePath := filepath.Join(dir, agentServerDetachReleaseDonePath)
	reportPath := filepath.Join(dir, agentServerDetachReportPath)

	require.NoError(t, AgentServerStartUp(identifier))
	AgentServerDetach()
	AgentServerJoin()

	// AgentServerShutDown after Detach would close ZMQ sockets the detached
	// native thread is still using. Skip orderly shutdown; the helper
	// process exits and the OS reclaims the thread.
	releaseSignaled := false
	signalReleaseDone := func() {
		if releaseSignaled {
			return
		}
		releaseSignaled = true
		_ = os.WriteFile(releaseDonePath, []byte("done"), 0o600)
	}
	defer signalReleaseDone()

	require.NoError(t, os.WriteFile(serverReadyPath, []byte("ready"), 0o600))
	waitAgentServerDetachFile(t, clientConnectedPath, agentServerDetachFileWait)

	// The server process owns no client or resource handles; any native
	// traffic is in the separate client process.
	live := liveNativeObjects.Load()
	require.Zero(t, live)

	// Intercept library unloading so a failed guard cannot dlclose code that
	// the detached native thread may still be executing.
	unloadAttempted := false
	actualShutdown := shutdownNativeLibraries
	shutdownNativeLibraries = func() error {
		unloadAttempted = true
		return nil
	}
	defer func() { shutdownNativeLibraries = actualShutdown }()

	err := Release()
	report := agentServerDetachReport{
		LiveNativeObjects: live,
		UnloadAttempted:   unloadAttempted,
	}
	if err != nil {
		report.ReleaseError = err.Error()
		report.ReleaseIsLibraryInUse = errors.Is(err, ErrLibraryInUse)
	}
	// Persist the outcome before any further cleanup so the parent can
	// assert the guard even if this helper later dies abruptly.
	writeAgentServerDetachReport(t, reportPath, report)
	// Unblock the client only after Release returns so it stays connected
	// through the guard check.
	signalReleaseDone()
	t.Logf("Release after Detach and Join returned %v; unload attempted: %t",
		err, unloadAttempted)
}

// runAgentServerDetachClientHelper connects to the detached Agent Server with
// the public Go wrappers and stays connected until the server has finished its
// Release check.
func runAgentServerDetachClientHelper(t *testing.T) {
	t.Helper()

	require.True(t, IsInited())

	dir := os.Getenv(agentServerDetachDirEnv)
	identifier := os.Getenv(agentServerDetachIDEnv)
	require.NotEmpty(t, dir)
	require.NotEmpty(t, identifier)

	serverReadyPath := filepath.Join(dir, agentServerDetachServerReadyPath)
	clientConnectedPath := filepath.Join(dir, agentServerDetachClientConnectedPath)
	releaseDonePath := filepath.Join(dir, agentServerDetachReleaseDonePath)

	waitAgentServerDetachFile(t, serverReadyPath, agentServerDetachFileWait)

	client, err := NewAgentClient(WithIdentifier(identifier))
	require.NoError(t, err)
	connected := false
	var res *Resource
	t.Cleanup(func() {
		if connected {
			_ = client.Disconnect()
		}
		_ = client.Destroy()
		if res != nil {
			_ = res.Destroy()
		}
	})

	require.NoError(t, client.SetTimeout(5*time.Second))
	res, err = NewResource()
	require.NoError(t, err)
	require.NoError(t, client.BindResource(res))
	require.NoError(t, client.Connect())
	connected = true
	require.True(t, client.Connected())
	require.True(t, client.Alive())

	// Handshake succeeded; tell the server it may call Release while this
	// client remains connected.
	require.NoError(t, os.WriteFile(clientConnectedPath, []byte("connected"), 0o600))
	waitAgentServerDetachFile(t, releaseDonePath, agentServerDetachFileWait)
}

func writeAgentServerDetachReport(t *testing.T, path string, report agentServerDetachReport) {
	t.Helper()

	data, err := json.Marshal(report)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func readAgentServerDetachReport(t *testing.T, path, logPath string) agentServerDetachReport {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "missing release report; server helper output:\n%s",
		readAgentServerDetachLog(logPath))
	var report agentServerDetachReport
	require.NoError(t, json.Unmarshal(data, &report))
	return report
}

// agentServerDetachChild tracks one helper process and its log path.
type agentServerDetachChild struct {
	role   string
	done   chan error
	exited bool
}

func startAgentServerDetachChild(t *testing.T, role, identifier, dir, logPath string) *agentServerDetachChild {
	t.Helper()

	logFile, err := os.Create(logPath)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), agentServerDetachProcWait+2*time.Second)
	cmd := exec.CommandContext(ctx, os.Args[0],
		"-test.run=^TestAgentServerDetachJoinKeepsReleaseBlocked$",
		"-test.count=1",
		"-test.v",
	)
	cmd.Env = append(os.Environ(),
		agentServerDetachRoleEnv+"="+role,
		agentServerDetachIDEnv+"="+identifier,
		agentServerDetachDirEnv+"="+dir,
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	require.NoError(t, cmd.Start())
	require.NoError(t, logFile.Close())

	child := &agentServerDetachChild{
		role: role,
		done: make(chan error, 1),
	}
	go func() {
		defer cancel()
		child.done <- cmd.Wait()
	}()

	t.Cleanup(func() {
		if child.exited {
			return
		}
		_ = cmd.Process.Kill()
		select {
		case <-child.done:
		case <-time.After(agentServerDetachKillWait):
			t.Errorf("%s helper process did not exit after kill", role)
		}
	})
	return child
}

func (c *agentServerDetachChild) wait(t *testing.T) error {
	t.Helper()

	select {
	case err := <-c.done:
		c.exited = true
		return err
	case <-time.After(agentServerDetachProcWait):
		return errors.New(c.role + " helper timed out")
	}
}

func waitAgentServerDetachFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(agentServerDetachPollPeriod)
	defer tick.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", filepath.Base(path))
		case <-tick.C:
		}
	}
}

func newAgentServerDetachIdentifier() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return "detach-release-" + hex.EncodeToString(buf[:]), nil
}

func readAgentServerDetachLog(path string) string {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err.Error()
	}
	return string(data)
}
