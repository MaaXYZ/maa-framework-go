package maa

import (
	"testing"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// newValidationAgentClient returns a stubbed client and captures SetTimeout calls.
func newValidationAgentClient(t *testing.T) (client *AgentClient, setTimeoutCalls func() []int64) {
	t.Helper()

	stubAgentClientCreation(t)
	client, err := NewAgentClient(WithIdentifier("validation"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Destroy()) })

	var calls []int64
	replaceNativeForTest(t, &native.MaaAgentClientSetTimeout, func(_ uintptr, milliseconds int64) bool {
		calls = append(calls, milliseconds)
		return true
	})
	return client, func() []int64 { return calls }
}

func TestAgentClient_SetTimeoutRounding(t *testing.T) {
	for _, tc := range []struct {
		name       string
		duration   time.Duration
		wantMillis int64
	}{
		{"ExplicitZeroPassesThrough", 0, 0},
		{"OneNanosecondRoundsUpToOne", 1 * time.Nanosecond, 1},
		{"SubMillisecondRoundsUpToOne", 999 * time.Microsecond, 1},
		{"ExactMillisecondStaysOne", time.Millisecond, 1},
		{"MillisecondFractionTruncatesToOne", 1500 * time.Microsecond, 1},
		{"TwoMillisecondsStayTwo", 2 * time.Millisecond, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, setTimeoutCalls := newValidationAgentClient(t)

			require.NoError(t, client.SetTimeout(tc.duration))
			require.Equal(t, []int64{tc.wantMillis}, setTimeoutCalls())
		})
	}
}

func TestAgentClient_ValidationSentinels(t *testing.T) {
	t.Run("NegativeTimeoutRejectedBeforeNative", func(t *testing.T) {
		client, setTimeoutCalls := newValidationAgentClient(t)
		require.ErrorIs(t, client.SetTimeout(-time.Second), ErrInvalidTimeout)
		require.Empty(t, setTimeoutCalls(), "negative timeout must not reach native SetTimeout")
	})

	t.Run("BindResourceNil", func(t *testing.T) {
		client, _ := newValidationAgentClient(t)
		require.ErrorIs(t, client.BindResource(nil), ErrInvalidResource)
	})

	t.Run("RegisterControllerSinkZeroValue", func(t *testing.T) {
		client, _ := newValidationAgentClient(t)
		require.ErrorIs(t, client.RegisterControllerSink(Controller{}), ErrInvalidController)
	})

	t.Run("RegisterTaskerSinkZeroValue", func(t *testing.T) {
		client, _ := newValidationAgentClient(t)
		require.ErrorIs(t, client.RegisterTaskerSink(Tasker{}), ErrInvalidTasker)
	})

	t.Run("NilClient", func(t *testing.T) {
		var client *AgentClient
		require.NoError(t, client.Destroy())
		_, err := client.Identifier()
		require.ErrorIs(t, err, ErrInvalidAgentClient)
	})
}
