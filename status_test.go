package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatus_Invalid(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: true,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: false,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: false,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: false,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Invalid()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_Pending(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: false,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: true,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: false,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: false,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Pending()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_Running(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: false,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: false,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: true,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: false,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Running()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_Success(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: false,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: false,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: false,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: true,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Success()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_Failed(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: false,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: false,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: false,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: false,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Failure()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_Done(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect bool
	}{
		{
			name:   "IsStatusInvalid",
			status: StatusInvalid,
			expect: false,
		},
		{
			name:   "IsStatusPending",
			status: StatusPending,
			expect: false,
		},
		{
			name:   "IsStatusRunning",
			status: StatusRunning,
			expect: false,
		},
		{
			name:   "IsStatusSuccess",
			status: StatusSuccess,
			expect: true,
		},
		{
			name:   "IsStatusFailed",
			status: StatusFailure,
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.Done()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestStatus_ConstantValues(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		value  int32
	}{
		{
			name:   "StatusInvalid",
			status: StatusInvalid,
			value:  0,
		},
		{
			name:   "StatusPending",
			status: StatusPending,
			value:  1000,
		},
		{
			name:   "StatusRunning",
			status: StatusRunning,
			value:  2000,
		},
		{
			name:   "StatusSuccess",
			status: StatusSuccess,
			value:  3000,
		},
		{
			name:   "StatusFailure",
			status: StatusFailure,
			value:  4000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.EqualValues(t, tc.value, tc.status)
		})
	}
}

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		expect string
	}{
		{
			name:   "StatusInvalid",
			status: StatusInvalid,
			expect: "invalid",
		},
		{
			name:   "StatusPending",
			status: StatusPending,
			expect: "pending",
		},
		{
			name:   "StatusRunning",
			status: StatusRunning,
			expect: "running",
		},
		{
			name:   "StatusSuccess",
			status: StatusSuccess,
			expect: "success",
		},
		{
			name:   "StatusFailure",
			status: StatusFailure,
			expect: "failure",
		},
		{
			name:   "UnknownValue",
			status: Status(9999),
			expect: "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.status.String()
			require.Equal(t, tc.expect, got)
		})
	}

	t.Run("UnknownValueDoesNotSatisfyInvalid", func(t *testing.T) {
		unknown := Status(9999)
		require.Equal(t, "invalid", unknown.String())
		require.False(t, unknown.Invalid())
		require.False(t, unknown.Done())
	})
}
