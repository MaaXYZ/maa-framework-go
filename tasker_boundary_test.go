package maa

import (
	"fmt"
	"math"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

// boundaryStatusController is a BlankController with explicit connection state.
// BlankController reports Connected unconditionally, so testing a bound but
// disconnected controller requires observable connection state.
type boundaryStatusController struct {
	BlankController
	connected atomic.Bool
}

func (c *boundaryStatusController) Connected() bool {
	return c.connected.Load()
}

func (c *boundaryStatusController) Connect() bool {
	c.connected.Store(true)
	return true
}

func TestTasker_GetTaskDetail_ZeroNodesBoundary(t *testing.T) {
	const (
		taskID = int64(9412)
		entry  = "BoundaryZeroNodesEntry"
	)
	cases := []struct {
		name   string
		status Status
	}{
		{name: "Pending", status: StatusPending},
		{name: "Running", status: StatusRunning},
		{name: "Success", status: StatusSuccess},
		{name: "Failure", status: StatusFailure},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tasker := createTasker(t)
			t.Cleanup(func() { require.NoError(t, tasker.Destroy()) })

			var calls atomic.Int64
			replaceNativeForTest(t, &native.MaaTaskerGetTaskDetail,
				func(_ uintptr, gotTaskID int64, entryBuffer uintptr, nodeIDs uintptr, size *uint64, status *int32) bool {
					calls.Add(1)
					require.Equal(t, taskID, gotTaskID)
					require.Zero(t, nodeIDs, "a zero-count query must not pass a node array")
					// Only provided outputs are written: an implementation that
					// passes no entry buffer and no status pointer fails the
					// assertions below instead of dereferencing null.
					if entryBuffer != 0 {
						buffer.NewStringBufferByHandle(entryBuffer).Set(entry)
					}
					if status != nil {
						*status = int32(tc.status)
					}
					*size = 0
					return true
				})

			detail, err := tasker.GetTaskDetail(taskID)
			require.NoError(t, err)
			require.NotNil(t, detail)
			require.Equal(t, taskID, detail.ID)
			require.Equal(t, entry, detail.Entry)
			require.Equal(t, tc.status, detail.Status)
			require.Nil(t, detail.Nodes)
			require.EqualValues(t, 1, calls.Load(), "a zero-count query must make exactly one native call")
		})
	}
}

func TestTasker_GetTaskDetail_NonZeroNodesBoundary(t *testing.T) {
	const (
		taskID      = int64(6270)
		entryFirst  = "BoundaryFirstEntry"
		entryLatest = "BoundaryLatestEntry"
	)

	t.Run("second query shrinks the returned size", func(t *testing.T) {
		tasker := createTasker(t)
		t.Cleanup(func() { require.NoError(t, tasker.Destroy()) })

		var calls atomic.Int64
		replaceNativeForTest(t, &native.MaaTaskerGetTaskDetail,
			func(_ uintptr, _ int64, entryBuffer uintptr, _ uintptr, size *uint64, status *int32) bool {
				call := calls.Add(1)
				if entryBuffer != 0 {
					if call == 1 {
						buffer.NewStringBufferByHandle(entryBuffer).Set(entryFirst)
					} else {
						buffer.NewStringBufferByHandle(entryBuffer).Set(entryLatest)
					}
				}
				if status != nil {
					if call == 1 {
						*status = int32(StatusRunning)
					} else {
						*status = int32(StatusFailure)
					}
				}
				// Report a shrink without touching node memory: the returned
				// size, not the previously allocated length, bounds traversal.
				if call == 1 {
					*size = 2
				} else {
					*size = 0
				}
				return true
			})

		var detail *TaskDetail
		var err error
		require.NotPanics(t, func() { detail, err = tasker.GetTaskDetail(taskID) })
		require.NoError(t, err)
		require.NotNil(t, detail)
		require.Equal(t, taskID, detail.ID)
		require.Equal(t, entryLatest, detail.Entry)
		require.Equal(t, StatusFailure, detail.Status)
		require.Len(t, detail.Nodes, 0)
		require.EqualValues(t, 2, calls.Load())
	})

	t.Run("second query fails", func(t *testing.T) {
		tasker := createTasker(t)
		t.Cleanup(func() { require.NoError(t, tasker.Destroy()) })

		var calls atomic.Int64
		replaceNativeForTest(t, &native.MaaTaskerGetTaskDetail,
			func(_ uintptr, _ int64, entryBuffer uintptr, _ uintptr, size *uint64, _ *int32) bool {
				if calls.Add(1) == 1 {
					if entryBuffer != 0 {
						buffer.NewStringBufferByHandle(entryBuffer).Set(entryFirst)
					}
					*size = 2
					return true
				}
				return false
			})

		detail, err := tasker.GetTaskDetail(taskID)
		require.Nil(t, detail)
		require.ErrorContains(t, err, fmt.Sprintf("taskId %d", taskID))
		require.EqualValues(t, 2, calls.Load())
	})
}

func TestTasker_GetTaskDetailRealNativeBoundary(t *testing.T) {
	const entry = "BoundaryAbsentEntry"

	res := createResource(t)
	tasker := createTasker(t)
	t.Cleanup(func() {
		// The tasker retains the resource binding until it is destroyed.
		require.NoError(t, tasker.Destroy())
		require.NoError(t, res.Destroy())
	})
	require.NoError(t, tasker.BindResource(res))

	job, err := tasker.PostTask(entry)
	require.NoError(t, err)
	require.True(t, job.Wait().Failure(), "an absent entry must fail")

	detail, err := job.GetDetail()
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.Equal(t, entry, detail.Entry)
	require.Equal(t, StatusFailure, detail.Status)
	require.Len(t, detail.Nodes, 0)

	const invalidTaskID = int64(math.MaxInt64)
	invalid, err := tasker.GetTaskDetail(invalidTaskID)
	require.Nil(t, invalid)
	require.ErrorContains(t, err, fmt.Sprintf("taskId %d", invalidTaskID))
}

func TestTasker_InitializedBoundary(t *testing.T) {
	tasker := createTasker(t)
	res := createResource(t)
	impl := &boundaryStatusController{}
	controller, err := NewCustomController(impl)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tasker.Destroy())
		require.NoError(t, controller.Destroy())
		require.NoError(t, res.Destroy())
	})

	require.False(t, tasker.Initialized(), "a tasker without a resource is not initialized")

	require.NoError(t, tasker.BindResource(res))
	require.True(t, tasker.Initialized(), "a fresh resource is valid before any bundle load")

	missing := filepath.Join(t.TempDir(), "absent-bundle")
	bundle, err := res.PostBundle(missing)
	require.NoError(t, err)
	require.True(t, bundle.Wait().Failure(), "loading an absent bundle must fail")
	require.False(t, tasker.Initialized(), "a failed bundle load invalidates the resource")

	require.NoError(t, res.Clear())
	require.True(t, tasker.Initialized(), "Clear restores resource validity")

	require.NoError(t, tasker.BindController(controller))
	require.False(t, tasker.Initialized(), "a bound disconnected controller keeps the tasker uninitialized")

	connect, err := controller.PostConnect()
	require.NoError(t, err)
	require.True(t, connect.Wait().Success())
	require.True(t, tasker.Initialized(), "a connected controller initializes the tasker")

	require.NoError(t, tasker.Destroy())
	require.False(t, tasker.Initialized(), "a destroyed tasker is not initialized")
	require.NoError(t, tasker.Destroy())

	require.NoError(t, controller.Destroy())
	require.NoError(t, res.Destroy())
}
