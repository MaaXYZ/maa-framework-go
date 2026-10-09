package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAgentClient_MethodsAfterDestroy pins that destroying a client closes its
// handle state: subsequent calls fail with ErrClosed before reaching native
// code, and the boolean probes report false instead of panicking.
func TestAgentClient_MethodsAfterDestroy(t *testing.T) {
	client, err := NewAgentClient(WithIdentifier("post-destroy-client"))
	require.NoError(t, err)

	// Sanity: the client works before destruction.
	identifier, err := client.Identifier()
	require.NoError(t, err)
	require.Equal(t, "post-destroy-client", identifier)

	require.NoError(t, client.Destroy())

	_, err = client.Identifier()
	require.ErrorIs(t, err, ErrClosed)
	require.ErrorIs(t, client.SetTimeout(0), ErrClosed)
	require.ErrorIs(t, client.Connect(), ErrClosed)
	require.False(t, client.Connected())
	require.False(t, client.Alive())

	// Destroy stays a nil-error no-op on an already-closed client.
	require.NoError(t, client.Destroy())
}
