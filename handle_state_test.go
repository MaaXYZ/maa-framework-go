package maa

import (
	"testing"

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
