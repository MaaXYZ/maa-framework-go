package maa

import (
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestTasker_DestroyRejectsPendingJobDuringCallbackSetup(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)

	oldPost := native.MaaTaskerPostTask
	oldStatus := native.MaaTaskerStatus
	oldGetTasker := native.MaaContextGetTasker
	defer func() {
		native.MaaTaskerPostTask = oldPost
		native.MaaTaskerStatus = oldStatus
		native.MaaContextGetTasker = oldGetTasker
	}()

	var status atomic.Int32
	status.Store(int32(StatusRunning))
	native.MaaTaskerPostTask = func(uintptr, string, string) int64 { return 42 }
	native.MaaTaskerStatus = func(uintptr, int64) int32 { return status.Load() }
	lookupStarted := make(chan struct{})
	finishLookup := make(chan struct{})
	native.MaaContextGetTasker = func(uintptr) uintptr {
		close(lookupStarted)
		<-finishLookup
		return tasker.handle
	}

	job := tasker.PostTask("entry")
	require.NoError(t, job.Error())
	callback := make(chan *Context, 1)
	go func() { callback <- newCallbackContext(123) }()
	<-lookupStarted
	firstClose := tasker.Destroy()
	close(finishLookup)
	ctx := <-callback
	callbackClose := tasker.Destroy()
	if ctx != nil {
		ctx.invalidate()
	}
	status.Store(int32(StatusSuccess))
	finalClose := tasker.Destroy()

	require.ErrorIs(t, firstClose, ErrInUse)
	require.NotNil(t, ctx)
	require.ErrorIs(t, callbackClose, ErrInCallback)
	require.NoError(t, finalClose)
}

func TestResource_DestroyRejectsPendingJob(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	oldPost := native.MaaResourcePostBundle
	oldStatus := native.MaaResourceStatus
	defer func() {
		native.MaaResourcePostBundle = oldPost
		native.MaaResourceStatus = oldStatus
	}()
	var status atomic.Int32
	status.Store(int32(StatusPending))
	native.MaaResourcePostBundle = func(uintptr, string) int64 { return 43 }
	native.MaaResourceStatus = func(uintptr, int64) int32 { return status.Load() }

	job := res.PostBundle("unused")
	require.NoError(t, job.Error())
	firstClose := res.Destroy()
	status.Store(int32(StatusSuccess))
	finalClose := res.Destroy()
	require.ErrorIs(t, firstClose, ErrInUse)
	require.NoError(t, finalClose)
}

func TestController_DestroyRejectsRunningJob(t *testing.T) {
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	oldPost := native.MaaControllerPostConnection
	oldStatus := native.MaaControllerStatus
	defer func() {
		native.MaaControllerPostConnection = oldPost
		native.MaaControllerStatus = oldStatus
	}()
	var status atomic.Int32
	status.Store(int32(StatusRunning))
	native.MaaControllerPostConnection = func(uintptr) int64 { return 44 }
	native.MaaControllerStatus = func(uintptr, int64) int32 { return status.Load() }

	job := ctrl.PostConnect()
	require.NoError(t, job.Error())
	firstClose := ctrl.Destroy()
	status.Store(int32(StatusFailure))
	finalClose := ctrl.Destroy()
	require.ErrorIs(t, firstClose, ErrInUse)
	require.NoError(t, finalClose)
}

func TestCallbackContext_RejectsClosedTaskerDuringSetup(t *testing.T) {
	state := &taskerState{handleState: newHandleState(45, nil)}
	taskerStates.Store(uintptr(45), state)
	defer taskerStates.Delete(uintptr(45))
	require.NoError(t, state.close())
	oldGetTasker := native.MaaContextGetTasker
	native.MaaContextGetTasker = func(uintptr) uintptr { return 45 }
	defer func() { native.MaaContextGetTasker = oldGetTasker }()

	require.Nil(t, newCallbackContext(123))
}

func TestCustomAction_NilTaskerDoesNotPanic(t *testing.T) {
	id := registerCustomAction(CustomActionFunc(func(*Context, *CustomActionArg) bool {
		t.Fatal("custom action ran without a tasker")
		return true
	}))
	defer unregisterCustomAction(id)
	oldGetTasker := native.MaaContextGetTasker
	native.MaaContextGetTasker = func(uintptr) uintptr { return 0 }
	defer func() { native.MaaContextGetTasker = oldGetTasker }()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("custom action panicked with a nil tasker: %v", recovered)
		}
	}()
	require.Zero(t, _MaaCustomActionCallbackAgent(123, 1, nil, nil, nil, 7, 0, uintptr(id)))
}
