package maa

import (
	"sync"
	"sync/atomic"
	"testing"

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
	require.NoError(t, state.close())
	require.ErrorIs(t, func() error { _, _, err := state.begin(); return err }(), ErrClosed)
	require.Zero(t, destroyed)
	done()
	require.Equal(t, 1, destroyed)
	require.NoError(t, state.close())
	require.Equal(t, 1, destroyed)
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
	require.Equal(t, res2.handle, tasker.GetResource().handle)
	require.Equal(t, ctrl2.handle, tasker.GetController().handle)
	require.NoError(t, res1.Destroy())
	require.NoError(t, ctrl1.Destroy())
	_, err = borrowedRes.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	_, err = borrowedCtrl.GetInfo()
	require.ErrorIs(t, err, ErrClosed)
	require.ErrorIs(t, res2.Destroy(), ErrBound)
	require.ErrorIs(t, ctrl2.Destroy(), ErrBound)

	require.NoError(t, tasker.Destroy())
	require.NoError(t, tasker.Destroy())
	require.Nil(t, tasker.GetResource())
	require.ErrorIs(t, tasker.PostTask("unused").Error(), ErrClosed)
	require.NoError(t, res2.Destroy())
	require.NoError(t, ctrl2.Destroy())
	require.NoError(t, res2.Destroy())
	require.NoError(t, ctrl2.Destroy())
	require.ErrorIs(t, res2.PostBundle("unused").Error(), ErrClosed)
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
	require.ErrorIs(t, tasker.Destroy(), ErrBorrowed)
	require.ErrorIs(t, ctrl.Destroy(), ErrBorrowed)
	require.ErrorIs(t, res.Destroy(), ErrBorrowed)
	require.Zero(t, tasker.AddSink(nil))
	require.Zero(t, ctrl.AddSink(nil))
	require.Zero(t, res.AddSink(nil))
	ctx.invalidate()
	require.False(t, tasker.Initialized())
	_, err := ctrl.GetInfo()
	require.ErrorIs(t, err, ErrClosed)
	_, err = res.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	require.Nil(t, clone.GetTasker())
	require.Nil(t, clone.Clone())
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
		client.Destroy()
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
	client.Destroy()
	client.Destroy()
	require.NoError(t, tasker.Destroy())
	require.NoError(t, ctrl.Destroy())
	require.NoError(t, res.Destroy())
}
