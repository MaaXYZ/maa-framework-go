package maa

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// stubAgentClientCreation replaces the native creation and destruction entry
// points with counters. Successful creations return a fake nonzero handle, so
// the returned client must be destroyed in cleanup; the destroy stub keeps that
// destruction away from the native library.
func stubAgentClientCreation(t *testing.T) (createV2 func() int, createTcp func() int, identifiers func() []string) {
	t.Helper()

	var v2Calls, tcpCalls int
	var v2Identifiers []string
	replaceNativeForTest(t, &native.MaaAgentClientCreateV2, func(identifier uintptr) uintptr {
		v2Calls++
		v2Identifiers = append(v2Identifiers, buffer.NewStringBufferByHandle(identifier).Get())
		return 12345
	})
	replaceNativeForTest(t, &native.MaaAgentClientCreateTcp, func(port uint16) uintptr {
		tcpCalls++
		return 23456
	})
	replaceNativeForTest(t, &native.MaaAgentClientDestroy, func(uintptr) {})
	return func() int { return v2Calls },
		func() int { return tcpCalls },
		func() []string { return v2Identifiers }
}

func TestAgentClient_CreationModesAndOptionPriority(t *testing.T) {
	for _, tc := range []struct {
		name      string
		opts      []AgentClientOption
		wantV2    int
		wantTcp   int
		wantIdent string
	}{
		{name: "NoOptionDefaultsToIdentifier", wantV2: 1, wantTcp: 0, wantIdent: ""},
		{name: "IdentifierOnly", opts: []AgentClientOption{WithIdentifier("mode-id")}, wantV2: 1, wantTcp: 0, wantIdent: "mode-id"},
		{name: "TcpPortOnly", opts: []AgentClientOption{WithTcpPort(5555)}, wantV2: 0, wantTcp: 1},
		{name: "IdentifierThenTcpPortUsesTcp", opts: []AgentClientOption{WithIdentifier("mode-id"), WithTcpPort(5555)}, wantV2: 0, wantTcp: 1},
		{name: "TcpPortThenIdentifierUsesIdentifier", opts: []AgentClientOption{WithTcpPort(5555), WithIdentifier("mode-id")}, wantV2: 1, wantTcp: 0, wantIdent: "mode-id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			createV2, createTcp, identifiers := stubAgentClientCreation(t)

			client, err := NewAgentClient(tc.opts...)
			require.NoError(t, err)
			require.NotNil(t, client)
			t.Cleanup(func() { require.NoError(t, client.Destroy()) })

			require.Equal(t, tc.wantV2, createV2(), "MaaAgentClientCreateV2 call count")
			require.Equal(t, tc.wantTcp, createTcp(), "MaaAgentClientCreateTcp call count")
			if tc.wantV2 > 0 {
				require.Equal(t, []string{tc.wantIdent}, identifiers(), "identifier passed to MaaAgentClientCreateV2")
			}
		})
	}
}

func TestAgentClient_TcpCreationForwardsPort(t *testing.T) {
	var ports []uint16
	replaceNativeForTest(t, &native.MaaAgentClientCreateTcp, func(port uint16) uintptr { ports = append(ports, port); return 23456 })
	replaceNativeForTest(t, &native.MaaAgentClientDestroy, func(uintptr) {})

	client, err := NewAgentClient(WithTcpPort(5555))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Destroy()) })

	require.Equal(t, []uint16{5555}, ports)
}

func TestAgentClient_CreateV2FailureReturnsError(t *testing.T) {
	replaceNativeForTest(t, &native.MaaAgentClientCreateV2, func(uintptr) uintptr { return 0 })

	client, err := NewAgentClient(WithIdentifier("unused"))
	require.Error(t, err)
	require.Nil(t, client)
}
