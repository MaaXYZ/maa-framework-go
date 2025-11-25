package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJob_Status(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     Status
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: StatusInvalid,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: StatusPending,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: StatusRunning,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: StatusSuccess,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: StatusFailure,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Status()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Invalid(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: true,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: false,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: false,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: false,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Invalid()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Pending(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: false,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: true,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: false,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: false,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Pending()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Running(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: false,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: false,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: true,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: false,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Running()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Success(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: false,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: false,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: false,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: true,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Success()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Failure(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: false,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: false,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: false,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: false,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Failure()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Done(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "IsStatusInvalid",
			id:   1,
			statusFunc: func(id int64) Status {
				return StatusInvalid
			},
			expect: false,
		},
		{
			name: "IsStatusPending",
			id:   2,
			statusFunc: func(id int64) Status {
				return StatusPending
			},
			expect: false,
		},
		{
			name: "IsStatusRunning",
			id:   3,
			statusFunc: func(id int64) Status {
				return StatusRunning
			},
			expect: false,
		},
		{
			name: "IsStatusSuccess",
			id:   4,
			statusFunc: func(id int64) Status {
				return StatusSuccess
			},
			expect: true,
		},
		{
			name: "IsStatusFailure",
			id:   5,
			statusFunc: func(id int64) Status {
				return StatusFailure
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := newJob(tc.id, tc.statusFunc, nil)
			got := job.Done()
			require.Equal(t, tc.expect, got)
		})
	}
}
