package maa

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestHandleState_CloseAfterActiveCall(t *testing.T) {
	var destroyed int
	state := newHandleState(123, func(handle uintptr) {
		require.Equal(t, uintptr(123), handle)
		destroyed++
	})
	_, done, err := state.begin()
	require.NoError(t, err)
	require.ErrorIs(t, state.close(), ErrInUse)
	require.NoError(t, state.check())
	require.Zero(t, destroyed)
	done()
	require.NoError(t, state.close())
	require.Equal(t, 1, destroyed)
	require.NoError(t, state.close())
	require.Equal(t, 1, destroyed)
}

func TestHandleState_ConcurrentCloseWaitsForCleanup(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	state := newHandleState(123, func(uintptr) {
		close(started)
		<-release
	})
	first := make(chan error, 1)
	go func() { first <- state.close() }()
	<-started
	second := make(chan error, 1)
	go func() { second <- state.close() }()
	select {
	case <-second:
		t.Fatal("repeat close returned before native cleanup completed")
	default:
	}
	close(release)
	require.NoError(t, <-first)
	require.NoError(t, <-second)
}

func TestHandleState_ConcurrentClose(t *testing.T) {
	var destroyed atomic.Int32
	state := newHandleState(123, func(uintptr) { destroyed.Add(1) })
	var callers sync.WaitGroup
	errs := make(chan error, 32)
	for range 32 {
		callers.Add(1)
		go func() {
			defer callers.Done()
			errs <- state.close()
		}()
	}
	callers.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, int32(1), destroyed.Load())
}

func TestHandleState_UntracksObservedCompletedJobs(t *testing.T) {
	state := newHandleState(123, func(uintptr) {})
	var firstStatus, secondStatus atomic.Int32
	firstStatus.Store(int32(StatusPending))
	secondStatus.Store(int32(StatusPending))
	state.jobStatus = func(_ uintptr, id int64) Status {
		if id == 1 {
			return Status(firstStatus.Load())
		}
		return Status(secondStatus.Load())
	}
	job := newJob(1,
		func(int64) Status { return Status(firstStatus.Load()) },
		func(int64) Status { return StatusSuccess },
		state,
	)
	require.Equal(t, StatusPending, job.Status())
	require.Len(t, state.jobs, 1)
	firstStatus.Store(int32(StatusSuccess))
	require.Equal(t, StatusSuccess, job.Status())
	require.Empty(t, state.jobs)

	job = newJob(2,
		func(int64) Status { return StatusPending },
		func(int64) Status {
			secondStatus.Store(int32(StatusSuccess))
			return StatusSuccess
		},
		state,
	)
	job.Wait()
	require.Empty(t, state.jobs)
	require.NoError(t, state.close())
}

func TestHandleState_PrunesDiscardedCompletedJobs(t *testing.T) {
	state := newHandleState(123, func(uintptr) {})
	var firstStatus, otherStatus atomic.Int32
	firstStatus.Store(int32(StatusRunning))
	otherStatus.Store(int32(StatusPending))
	state.jobStatus = func(_ uintptr, id int64) Status {
		if id == 1 {
			return Status(firstStatus.Load())
		}
		return Status(otherStatus.Load())
	}
	for id := int64(1); id <= 128; id++ {
		newJob(id, nil, nil, state)
	}
	otherStatus.Store(int32(StatusSuccess))
	require.Eventually(t, func() bool {
		state.mu.Lock()
		defer state.mu.Unlock()
		return len(state.jobs) == 1
	}, 3*time.Second, 10*time.Millisecond)
	state.mu.Lock()
	_, retained := state.jobs[1]
	state.mu.Unlock()
	require.True(t, retained)
	require.ErrorIs(t, state.close(), ErrInUse)
	firstStatus.Store(int32(StatusSuccess))
	require.Eventually(t, func() bool {
		state.mu.Lock()
		defer state.mu.Unlock()
		return len(state.jobs) == 0
	}, 3*time.Second, 10*time.Millisecond)
	require.NoError(t, state.close())
}

func TestTasker_BorrowedBindingsAndRebinding(t *testing.T) {
	res1, err := NewResource()
	require.NoError(t, err)
	res2, err := NewResource()
	require.NoError(t, err)
	ctrl1, err := NewBlankController()
	require.NoError(t, err)
	ctrl2, err := NewBlankController()
	require.NoError(t, err)
	tasker, err := NewTasker()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tasker.Destroy())
		require.NoError(t, res1.Destroy())
		require.NoError(t, res2.Destroy())
		require.NoError(t, ctrl1.Destroy())
		require.NoError(t, ctrl2.Destroy())
	})

	require.Nil(t, tasker.GetResource())
	require.Nil(t, tasker.GetController())
	require.NoError(t, tasker.BindResource(res1))
	require.NoError(t, tasker.BindController(ctrl1))
	borrowedRes := tasker.GetResource()
	borrowedCtrl := tasker.GetController()
	require.NotNil(t, borrowedRes)
	require.NotNil(t, borrowedCtrl)
	require.ErrorIs(t, borrowedRes.Destroy(), ErrBorrowed)
	require.ErrorIs(t, borrowedCtrl.Destroy(), ErrBorrowed)
	require.ErrorIs(t, res1.Destroy(), ErrBound)
	require.ErrorIs(t, ctrl1.Destroy(), ErrBound)

	require.NoError(t, tasker.BindResource(res2))
	require.NoError(t, tasker.BindController(ctrl2))
	require.NoError(t, tasker.BindResource(res2))
	require.NoError(t, tasker.BindController(ctrl2))
	require.Equal(t, res2.handle, tasker.GetResource().handle)
	require.Equal(t, ctrl2.handle, tasker.GetController().handle)
	require.ErrorIs(t, res1.Destroy(), ErrBound)
	require.ErrorIs(t, ctrl1.Destroy(), ErrBound)
	require.NoError(t, borrowedRes.state.check())
	require.NoError(t, borrowedCtrl.state.check())
	require.ErrorIs(t, res2.Destroy(), ErrBound)
	require.ErrorIs(t, ctrl2.Destroy(), ErrBound)

	require.NoError(t, tasker.Destroy())
	require.NoError(t, tasker.Destroy())
	require.Nil(t, tasker.GetResource())
	require.ErrorIs(t, tasker.PostTask("unused").Error(), ErrClosed)
	require.NoError(t, res1.Destroy())
	require.NoError(t, ctrl1.Destroy())
	_, err = borrowedRes.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	_, err = borrowedCtrl.GetInfo()
	require.ErrorIs(t, err, ErrClosed)
	require.NoError(t, res2.Destroy())
	require.NoError(t, ctrl2.Destroy())
	require.NoError(t, res2.Destroy())
	require.NoError(t, ctrl2.Destroy())
	require.ErrorIs(t, res2.PostBundle("unused").Error(), ErrClosed)
}

func TestTasker_RejectsBindingsWhileRunning(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)
	res, err := NewResource()
	require.NoError(t, err)
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tasker.Destroy())
		require.NoError(t, res.Destroy())
		require.NoError(t, ctrl.Destroy())
	}()

	oldRunning := native.MaaTaskerRunning
	native.MaaTaskerRunning = func(uintptr) bool { return true }
	defer func() { native.MaaTaskerRunning = oldRunning }()
	require.ErrorIs(t, tasker.BindResource(res), ErrTaskerRunning)
	require.ErrorIs(t, tasker.BindController(ctrl), ErrTaskerRunning)
	require.Nil(t, tasker.GetResource())
	require.Nil(t, tasker.GetController())
	require.NoError(t, res.state.check())
	require.NoError(t, ctrl.state.check())
}

func TestTasker_PostAndBindDoNotInterleave(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)
	res, err := NewResource()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tasker.Destroy())
		require.NoError(t, res.Destroy())
	}()

	oldPost := native.MaaTaskerPostTask
	oldRunning := native.MaaTaskerRunning
	started := make(chan struct{})
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	var running atomic.Bool
	native.MaaTaskerPostTask = func(uintptr, string, string) int64 {
		close(started)
		<-release
		return 0
	}
	native.MaaTaskerRunning = func(uintptr) bool { return running.Load() }
	defer func() {
		native.MaaTaskerPostTask = oldPost
		native.MaaTaskerRunning = oldRunning
	}()

	posted := make(chan struct{})
	go func() {
		tasker.PostTask("entry")
		close(posted)
	}()
	<-started
	bound := make(chan error, 1)
	go func() { bound <- tasker.BindResource(res) }()
	select {
	case <-bound:
		t.Fatal("bind completed while posting was still in progress")
	case <-time.After(20 * time.Millisecond):
	}
	running.Store(true)
	close(release)
	<-posted
	require.ErrorIs(t, <-bound, ErrTaskerRunning)
	require.Nil(t, tasker.GetResource())
}

type destroyOnMarshal struct {
	tasker *Tasker
	err    error
}

func (v *destroyOnMarshal) MarshalJSON() ([]byte, error) {
	v.err = v.tasker.Destroy()
	return []byte(`{}`), nil
}

func TestTasker_DestroyFromMarshalReturnsInUse(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)
	oldPost := native.MaaTaskerPostTask
	native.MaaTaskerPostTask = func(uintptr, string, string) int64 { return 0 }
	defer func() { native.MaaTaskerPostTask = oldPost }()

	value := &destroyOnMarshal{tasker: tasker}
	tasker.PostTask("entry", value)
	require.ErrorIs(t, value.err, ErrInUse)
	require.NoError(t, tasker.Destroy())
}

func TestEventBorrowDoesNotOwnHandle(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	defer func() { require.NoError(t, res.Destroy()) }()
	var borrowed *Resource
	handleResourceLoading(resourceSinkFunc(func(r *Resource) {
		borrowed = r
	}), res.handle, EventStatusStarting, []byte(`{"res_id":1}`))
	require.NotNil(t, borrowed)
	require.ErrorIs(t, borrowed.Destroy(), ErrBorrowed)
	require.NoError(t, res.Destroy())
	_, err = borrowed.GetHash()
	require.ErrorIs(t, err, ErrClosed)
}

func TestEventCallback_CannotDestroyOwner(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	var closeErr error
	handleResourceLoading(resourceSinkFunc(func(*Resource) {
		closeErr = res.Destroy()
	}), res.handle, EventStatusStarting, []byte(`{"res_id":1}`))
	require.ErrorIs(t, closeErr, ErrInCallback)
	require.NoError(t, res.Destroy())
}

func TestEventBorrow_ExternalHandleExpires(t *testing.T) {
	var borrowed *Resource
	handleResourceLoading(resourceSinkFunc(func(r *Resource) {
		borrowed = r
	}), 98765, EventStatusStarting, []byte(`{"res_id":1}`))
	require.NotNil(t, borrowed)
	require.ErrorIs(t, borrowed.Destroy(), ErrBorrowed)
	_, err := borrowed.GetHash()
	require.ErrorIs(t, err, ErrClosed)
}

type resourceSinkFunc func(*Resource)

func (f resourceSinkFunc) OnResourceLoading(res *Resource, _ EventStatus, _ ResourceLoadingDetail) {
	f(res)
}

func TestCallbackContextExpires(t *testing.T) {
	var captured *Context
	sink := &contextEventSinkAdapter{
		onNodePipelineNode: func(ctx *Context, _ EventStatus, _ NodePipelineNodeDetail) {
			captured = ctx
		},
	}
	handleNodePipelineNode(sink, 0, EventStatusStarting, []byte(`{}`))
	require.NotNil(t, captured)
	_, err := captured.GetNodeJSON("unused")
	require.ErrorIs(t, err, ErrClosed)
	require.Nil(t, captured.GetTasker())
	require.Nil(t, captured.Clone())
}

func TestCallbackContext_ExternalTaskerAndControllerExpire(t *testing.T) {
	oldGetTasker := native.MaaContextGetTasker
	oldGetController := native.MaaTaskerGetController
	oldGetResource := native.MaaTaskerGetResource
	oldClone := native.MaaContextClone
	native.MaaContextGetTasker = func(uintptr) uintptr { return 123 }
	native.MaaTaskerGetController = func(uintptr) uintptr { return 456 }
	native.MaaTaskerGetResource = func(uintptr) uintptr { return 457 }
	native.MaaContextClone = func(uintptr) uintptr { return 790 }
	defer func() {
		native.MaaContextGetTasker = oldGetTasker
		native.MaaTaskerGetController = oldGetController
		native.MaaTaskerGetResource = oldGetResource
		native.MaaContextClone = oldClone
	}()

	ctx := newCallbackContext(789)
	clone := ctx.Clone()
	require.NotNil(t, clone)
	tasker := ctx.GetTasker()
	require.NotNil(t, tasker)
	ctrl := tasker.GetController()
	require.NotNil(t, ctrl)
	res := tasker.GetResource()
	require.NotNil(t, res)
	localTasker, err := NewTasker()
	require.NoError(t, err)
	require.ErrorIs(t, localTasker.BindResource(res), ErrBorrowed)
	require.NoError(t, localTasker.Destroy())
	require.ErrorIs(t, tasker.Destroy(), ErrBorrowed)
	require.ErrorIs(t, ctrl.Destroy(), ErrBorrowed)
	require.ErrorIs(t, res.Destroy(), ErrBorrowed)
	require.Zero(t, tasker.AddSink(nil))
	require.Zero(t, ctrl.AddSink(nil))
	require.Zero(t, res.AddSink(nil))
	ctx.invalidate()
	require.False(t, tasker.Initialized())
	_, err = ctrl.GetInfo()
	require.ErrorIs(t, err, ErrClosed)
	_, err = res.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	require.Nil(t, clone.GetTasker())
	require.Nil(t, clone.Clone())
}

func TestCallbackContext_ActiveRunCompletesDuringInvalidation(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)
	oldGetTasker := native.MaaContextGetTasker
	oldRunTask := native.MaaContextRunTask
	oldGetDetail := native.MaaTaskerGetTaskDetail
	started := make(chan struct{})
	release := make(chan struct{})
	native.MaaContextGetTasker = func(uintptr) uintptr { return tasker.handle }
	native.MaaContextRunTask = func(uintptr, string, string) int64 {
		close(started)
		<-release
		return 42
	}
	native.MaaTaskerGetTaskDetail = func(_ uintptr, _ int64, _ uintptr, _ uintptr, size *uint64, _ *int32) bool {
		*size = 0
		return true
	}
	defer func() {
		native.MaaContextGetTasker = oldGetTasker
		native.MaaContextRunTask = oldRunTask
		native.MaaTaskerGetTaskDetail = oldGetDetail
		require.NoError(t, tasker.Destroy())
	}()

	ctx := newCallbackContext(789)
	result := make(chan error, 1)
	go func() {
		_, runErr := ctx.RunTask("entry")
		result <- runErr
	}()
	<-started
	invalidated := make(chan struct{})
	go func() {
		ctx.invalidate()
		close(invalidated)
	}()
	require.Eventually(t, func() bool {
		ctx.state.mu.Lock()
		defer ctx.state.mu.Unlock()
		return ctx.state.closed
	}, time.Second, time.Millisecond)
	close(release)
	require.NoError(t, <-result)
	<-invalidated
	require.Nil(t, ctx.GetTasker())
}

func TestAgentClient_HoldsRegisteredHandles(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	tasker, err := NewTasker()
	require.NoError(t, err)
	oldDestroy := native.MaaAgentClientDestroy
	oldBind := native.MaaAgentClientBindResource
	oldResSink := native.MaaAgentClientRegisterResourceSink
	oldCtrlSink := native.MaaAgentClientRegisterControllerSink
	oldTaskerSink := native.MaaAgentClientRegisterTaskerSink
	native.MaaAgentClientDestroy = func(uintptr) {}
	native.MaaAgentClientBindResource = func(uintptr, uintptr) bool { return true }
	native.MaaAgentClientRegisterResourceSink = func(uintptr, uintptr) bool { return true }
	native.MaaAgentClientRegisterControllerSink = func(uintptr, uintptr) bool { return true }
	native.MaaAgentClientRegisterTaskerSink = func(uintptr, uintptr) bool { return true }
	defer func() {
		native.MaaAgentClientDestroy = oldDestroy
		native.MaaAgentClientBindResource = oldBind
		native.MaaAgentClientRegisterResourceSink = oldResSink
		native.MaaAgentClientRegisterControllerSink = oldCtrlSink
		native.MaaAgentClientRegisterTaskerSink = oldTaskerSink
	}()
	client := newAgentClientByHandle(123)
	t.Cleanup(func() {
		require.NoError(t, client.Destroy())
		require.NoError(t, tasker.Destroy())
		require.NoError(t, ctrl.Destroy())
		require.NoError(t, res.Destroy())
	})

	require.NoError(t, client.BindResource(res))
	require.NoError(t, client.RegisterResourceSink(res))
	require.NoError(t, client.RegisterControllerSink(*ctrl))
	require.NoError(t, client.RegisterTaskerSink(*tasker))
	require.ErrorIs(t, res.Destroy(), ErrBound)
	require.ErrorIs(t, ctrl.Destroy(), ErrBound)
	require.ErrorIs(t, tasker.Destroy(), ErrBound)
	require.NoError(t, client.Destroy())
	require.NoError(t, client.Destroy())
	require.NoError(t, tasker.Destroy())
	require.NoError(t, ctrl.Destroy())
	require.NoError(t, res.Destroy())
}

func TestAgentClient_DestroyReportsActiveCall(t *testing.T) {
	oldDestroy := native.MaaAgentClientDestroy
	var destroyed int
	native.MaaAgentClientDestroy = func(uintptr) { destroyed++ }
	defer func() { native.MaaAgentClientDestroy = oldDestroy }()
	client := newAgentClientByHandle(123)
	_, done, err := client.state.begin()
	require.NoError(t, err)
	require.ErrorIs(t, client.Destroy(), ErrInUse)
	require.Zero(t, destroyed)
	done()
	require.NoError(t, client.Destroy())
	require.NoError(t, client.Destroy())
	require.Equal(t, 1, destroyed)
}
