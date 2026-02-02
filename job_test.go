package maa

import (
	"errors"
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

// TaskJob tests

func TestTaskJob_Error(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		job := newTaskJob(1, nil, nil, nil, nil, nil)
		require.NoError(t, job.Error())
	})

	t.Run("WithError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		job := newTaskJob(0, nil, nil, nil, nil, expectedErr)
		require.Equal(t, expectedErr, job.Error())
	})
}

func TestTaskJob_Status(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.Equal(t, StatusFailure, job.Status())
	})

	t.Run("NoError", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.Equal(t, StatusSuccess, job.Status())
	})
}

func TestTaskJob_Wait(t *testing.T) {
	t.Run("WithError_SkipsWait", func(t *testing.T) {
		waitCalled := false
		waitFunc := func(id int64) Status {
			waitCalled = true
			return StatusSuccess
		}
		job := newTaskJob(0, nil, waitFunc, nil, nil, errors.New("test error"))
		result := job.Wait()

		require.False(t, waitCalled, "waitFunc should not be called when error exists")
		require.Same(t, job, result, "Wait should return the same TaskJob instance")
	})

	t.Run("NoError_CallsWait", func(t *testing.T) {
		waitCalled := false
		waitFunc := func(id int64) Status {
			waitCalled = true
			return StatusSuccess
		}
		job := newTaskJob(1, nil, waitFunc, nil, nil, nil)
		result := job.Wait()

		require.True(t, waitCalled, "waitFunc should be called when no error exists")
		require.Same(t, job, result, "Wait should return the same TaskJob instance")
	})
}

func TestTaskJob_Invalid(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.False(t, job.Invalid(), "Invalid should return false when error exists (status is Failure)")
	})

	t.Run("NoError_StatusInvalid", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusInvalid }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Invalid())
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Invalid())
	})
}

func TestTaskJob_Pending(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.False(t, job.Pending(), "Pending should return false when error exists")
	})

	t.Run("NoError_StatusPending", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusPending }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Pending())
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Pending())
	})
}

func TestTaskJob_Running(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.False(t, job.Running(), "Running should return false when error exists")
	})

	t.Run("NoError_StatusRunning", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusRunning }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Running())
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Running())
	})
}

func TestTaskJob_Success(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.False(t, job.Success(), "Success should return false when error exists")
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Success())
	})

	t.Run("NoError_StatusFailure", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusFailure }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Success())
	})
}

func TestTaskJob_Failure(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.True(t, job.Failure(), "Failure should return true when error exists")
	})

	t.Run("NoError_StatusFailure", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusFailure }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Failure())
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Failure())
	})
}

func TestTaskJob_Done(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		job := newTaskJob(0, nil, nil, nil, nil, errors.New("test error"))
		require.True(t, job.Done(), "Done should return true when error exists (status is Failure)")
	})

	t.Run("NoError_StatusSuccess", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusSuccess }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Done())
	})

	t.Run("NoError_StatusFailure", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusFailure }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.True(t, job.Done())
	})

	t.Run("NoError_StatusPending", func(t *testing.T) {
		statusFunc := func(id int64) Status { return StatusPending }
		job := newTaskJob(1, statusFunc, nil, nil, nil, nil)
		require.False(t, job.Done())
	})
}

func TestTaskJob_GetDetail(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		job := newTaskJob(0, nil, nil, nil, nil, expectedErr)
		detail, err := job.GetDetail()
		require.Nil(t, detail)
		require.Equal(t, expectedErr, err)
	})

	t.Run("NoError_NilFunc", func(t *testing.T) {
		job := newTaskJob(1, nil, nil, nil, nil, nil)
		detail, err := job.GetDetail()
		require.Nil(t, detail)
		require.Error(t, err)
		require.Contains(t, err.Error(), "getTaskDetailFunc is nil")
	})

	t.Run("NoError_WithFunc", func(t *testing.T) {
		expectedDetail := &TaskDetail{ID: 1, Entry: "test"}
		getDetailFunc := func(id int64) (*TaskDetail, error) {
			return expectedDetail, nil
		}
		job := newTaskJob(1, nil, nil, getDetailFunc, nil, nil)
		detail, err := job.GetDetail()
		require.NoError(t, err)
		require.Equal(t, expectedDetail, detail)
	})
}

func TestTaskJob_OverridePipeline(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		job := newTaskJob(0, nil, nil, nil, nil, expectedErr)
		err := job.OverridePipeline("{}")
		require.Equal(t, expectedErr, err)
	})

	t.Run("NoError_NilFunc", func(t *testing.T) {
		job := newTaskJob(1, nil, nil, nil, nil, nil)
		err := job.OverridePipeline("{}")
		require.Error(t, err)
		require.Contains(t, err.Error(), "overridePipelineFunc is nil")
	})

	t.Run("NoError_WithFunc", func(t *testing.T) {
		overrideCalled := false
		overrideFunc := func(id int64, override any) error {
			overrideCalled = true
			return nil
		}
		job := newTaskJob(1, nil, nil, nil, overrideFunc, nil)
		err := job.OverridePipeline("{}")
		require.NoError(t, err)
		require.True(t, overrideCalled)
	})
}
