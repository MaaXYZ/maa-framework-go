//go:build !race

package maa

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

	// agentServerDetachParkWatch is how long the parent watches a parked helper
	// before terminating it. A helper that returned from its Go test instead of
	// parking exits within this window.
	agentServerDetachParkWatch = 250 * time.Millisecond
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
// wrappers and stays connected until after Release returns. The server helper
// parks once it has reported the outcome, and the parent asserts that report
// before terminating it, because exiting with the detached native service
// thread still running is not safe on Windows.
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

	// The server helper writes the report before it signals release completion,
	// so that signal also proves the report file is complete.
	waitAgentServerDetachRelease(t, filepath.Join(dir, agentServerDetachReleaseDonePath), server, client)

	// Assert the guard from the report while the server helper is still parked.
	report := readAgentServerDetachReport(t, filepath.Join(dir, agentServerDetachReportPath), serverLogPath)
	assertAgentServerDetachReleaseBlocked(t, report)

	// Release has returned and the report is assertable; let the client, which
	// stayed connected through Release, disconnect and exit.
	if clientErr := client.wait(t); clientErr != nil {
		t.Fatalf("client helper failed: %v\n%s", clientErr, agentServerDetachLogs(server, client))
	}

	// The server helper must still be parked. Had the helper returned from its
	// Go test, normal process teardown would have run with the detached native
	// service thread still alive, which is the failure this guards against.
	server.requireRunning(t)

	// The parent owns termination; the kill status is expected and is not a
	// helper failure.
	server.terminate(t)
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
	// native thread is still using. Skip orderly shutdown; the helper parks
	// below and the parent terminates the process instead.
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

	// Returning from this test would run normal process teardown while the
	// detached native service thread still owns live ZeroMQ sockets. On Windows
	// that teardown asserts with "Successful WSASTARTUP not yet performed", so
	// park until the parent, which now has the report, terminates this helper.
	parkAgentServerDetachServerHelper()
}

// parkAgentServerDetachServerHelper blocks the helper test forever without
// returning from it. Sleeping keeps a timer armed so the Go runtime cannot
// mistake the parked process for a deadlock; the parent terminates the process
// instead of letting the test return.
func parkAgentServerDetachServerHelper() {
	for {
		time.Sleep(time.Minute)
	}
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

// agentServerDetachChild tracks one helper process, its exit state, and its
// log path.
type agentServerDetachChild struct {
	role    string
	logPath string
	cmd     *exec.Cmd
	done    chan error
	exited  bool
	exitErr error
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
		role:    role,
		logPath: logPath,
		cmd:     cmd,
		done:    make(chan error, 1),
	}
	go func() {
		defer cancel()
		child.done <- cmd.Wait()
	}()

	t.Cleanup(func() {
		if child.killAndReap() {
			return
		}
		t.Errorf("%s helper process did not exit after kill", role)
	})
	return child
}

// recordExit stores the helper's wait result. Every receive from done must go
// through this so later checks see a consistent exit state.
func (c *agentServerDetachChild) recordExit(err error) {
	c.exited = true
	c.exitErr = err
}

// pollExit reports whether the helper has exited, draining a wait result that
// is already available.
func (c *agentServerDetachChild) pollExit() bool {
	if !c.exited {
		select {
		case err := <-c.done:
			c.recordExit(err)
		default:
		}
	}
	return c.exited
}

func (c *agentServerDetachChild) wait(t *testing.T) error {
	t.Helper()

	if c.pollExit() {
		return c.exitErr
	}
	select {
	case err := <-c.done:
		c.recordExit(err)
		return err
	case <-time.After(agentServerDetachProcWait):
		return errors.New(c.role + " helper timed out")
	}
}

// requireRunning fails when a helper that must stay parked has exited. The
// helper may only leave that state through the parent's kill, so an exit means
// normal process teardown ran while the detached native service thread was
// still live.
func (c *agentServerDetachChild) requireRunning(t *testing.T) {
	t.Helper()

	if !c.pollExit() {
		// Watch briefly so a helper that returned from its Go test right after
		// writing the report is still caught before the parent kills it.
		select {
		case err := <-c.done:
			c.recordExit(err)
		case <-time.After(agentServerDetachParkWatch):
		}
	}
	if !c.exited {
		return
	}
	t.Fatalf("%s helper exited instead of staying alive for the parent to terminate it: %v\n%s",
		c.role, c.exitErr, readAgentServerDetachLog(c.logPath))
}

// terminate kills a parked helper and reaps it. The helper can only leave the
// parked state through this kill, so the resulting nonzero exit status is
// expected and deliberately ignored.
func (c *agentServerDetachChild) terminate(t *testing.T) {
	t.Helper()

	if c.killAndReap() {
		return
	}
	t.Errorf("%s helper did not exit after the parent killed it\n%s",
		c.role, readAgentServerDetachLog(c.logPath))
}

// killAndReap stops the helper if it is still running and waits for it to be
// reaped, reporting whether the helper is known to have exited.
func (c *agentServerDetachChild) killAndReap() bool {
	if c.pollExit() {
		return true
	}
	if err := c.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return false
	}
	select {
	case err := <-c.done:
		c.recordExit(err)
		return true
	case <-time.After(agentServerDetachKillWait):
		return false
	}
}

// agentServerDetachLogs formats both helper logs for a parent failure message.
func agentServerDetachLogs(server, client *agentServerDetachChild) string {
	return fmt.Sprintf("server helper log:\n%s\nclient helper log:\n%s",
		readAgentServerDetachLog(server.logPath), readAgentServerDetachLog(client.logPath))
}

// waitAgentServerDetachRelease waits for the server helper to signal that
// Release returned. The report is written before that signal, so the signal
// also proves the report file is complete. Either helper exiting before the
// signal, or the timeout expiring, fails the parent with both logs.
func waitAgentServerDetachRelease(t *testing.T, path string, server, client *agentServerDetachChild) {
	t.Helper()

	deadline := time.NewTimer(agentServerDetachFileWait)
	defer deadline.Stop()
	tick := time.NewTicker(agentServerDetachPollPeriod)
	defer tick.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		select {
		case err := <-server.done:
			server.recordExit(err)
			// The signal may have appeared before the exit was observed; in
			// that case the report assertion and the parked-helper check give
			// the more precise diagnostics.
			if _, statErr := os.Stat(path); statErr == nil {
				return
			}
			t.Fatalf("server helper exited before signaling release completion: %v\n%s",
				err, agentServerDetachLogs(server, client))
		case err := <-client.done:
			client.recordExit(err)
			// The client exits as soon as the signal exists, so re-check the
			// file before treating this as a failure.
			if _, statErr := os.Stat(path); statErr == nil {
				return
			}
			t.Fatalf("client helper exited before the server signaled release completion: %v\n%s",
				err, agentServerDetachLogs(server, client))
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s from the server helper\n%s",
				filepath.Base(path), agentServerDetachLogs(server, client))
		case <-tick.C:
		}
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
