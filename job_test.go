package maa

import (
	"github.com/stretchr/testify/require"
	"testing"
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
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
			job := NewJob(tc.id, tc.statusFunc)
			got := job.Done()
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestJob_Wait(t *testing.T) {
	testCases := []struct {
		name       string
		id         int64
		statusFunc func(id int64) Status
		expect     bool
	}{
		{
			name: "WaitUntilSuccess",
			id:   1,
			statusFunc: func() func(id int64) Status {
				var count = 0
				return func(id int64) Status {
					if count < 10 {
						count++
						return StatusPending
					} else if count < 20 {
						count++
						return StatusRunning
					}
					return StatusSuccess
				}
			}(),
			expect: true,
		},
		{
			name: "WaitUntilFailure",
			id:   2,
			statusFunc: func() func(id int64) Status {
				var count = 0
				return func(id int64) Status {
					if count < 10 {
						count++
						return StatusPending
					} else if count < 20 {
						count++
						return StatusRunning
					}
					return StatusFailure
				}
			}(),
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job := NewJob(tc.id, tc.statusFunc)
			got := job.Wait()
			require.Equal(t, tc.expect, got)
		})
	}
}
