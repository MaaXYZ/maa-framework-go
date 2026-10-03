package maa

import (
	"sync/atomic"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestBorrowedBoundary_LocalCallbackResourceView(t *testing.T) {
	res := createResource(t)
	t.Cleanup(func() { require.NoError(t, res.Destroy()) })

	var borrowed *Resource
	handleResourceLoading(resourceSinkFunc(func(r *Resource) {
		borrowed = r
	}), res.handle, EventStatusStarting, []byte(`{"res_id":1}`))
	require.NotNil(t, borrowed)
	require.NotSame(t, res, borrowed)
	require.Same(t, res.state, borrowed.state, "a local view shares the owner's state")

	var hashCalls atomic.Int64
	replaceNativeForTest(t, &native.MaaResourceGetHash, func(uintptr, uintptr) bool {
		hashCalls.Add(1)
		return true
	})

	require.ErrorIs(t, borrowed.Destroy(), ErrBorrowed)
	_, err := borrowed.GetHash()
	require.NoError(t, err, "a rejected Destroy must not invalidate the borrowed view")
	require.EqualValues(t, 1, hashCalls.Load())

	require.NoError(t, res.Destroy())
	_, err = borrowed.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	require.EqualValues(t, 1, hashCalls.Load(), "a view closed with its owner must not reach native code")
}

func TestBorrowedBoundary_ContextBackedViewsExpire(t *testing.T) {
	var nativeCalls atomic.Int64
	count := func() { nativeCalls.Add(1) }

	replaceNativeForTest(t, &native.MaaContextGetTasker, func(uintptr) uintptr {
		count()
		return 123
	})
	replaceNativeForTest(t, &native.MaaContextClone, func(uintptr) uintptr {
		count()
		return 790
	})
	replaceNativeForTest(t, &native.MaaTaskerGetController, func(uintptr) uintptr {
		count()
		return 456
	})
	replaceNativeForTest(t, &native.MaaTaskerGetResource, func(uintptr) uintptr {
		count()
		return 457
	})
	replaceNativeForTest(t, &native.MaaTaskerInited, func(uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaTaskerGetTaskDetail, func(uintptr, int64, uintptr, uintptr, *uint64, *int32) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaControllerConnected, func(uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaControllerGetInfo, func(uintptr, uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaResourceGetHash, func(uintptr, uintptr) bool {
		count()
		return true
	})
	replaceNativeForTest(t, &native.MaaContextGetNodeData, func(uintptr, string, uintptr) bool {
		count()
		return true
	})

	ctx := newCallbackContext(789)
	require.NotNil(t, ctx)
	clone := ctx.Clone()
	require.NotNil(t, clone)
	tasker := ctx.GetTasker()
	require.NotNil(t, tasker)
	controller := tasker.GetController()
	require.NotNil(t, controller)
	resource := tasker.GetResource()
	require.NotNil(t, resource)

	before := nativeCalls.Load()
	require.EqualValues(t, 5, before, "view lookup must use the callback scope")

	ctx.invalidate()

	// Representative public methods report ErrClosed / false / nil after the
	// callback scope closes, and none of them may invoke native code.
	require.False(t, tasker.Initialized())
	detail, err := tasker.GetTaskDetail(7)
	require.Nil(t, detail)
	require.ErrorIs(t, err, ErrClosed)
	require.False(t, controller.Connected())
	_, err = controller.GetInfo()
	require.ErrorIs(t, err, ErrClosed)
	_, err = resource.GetHash()
	require.ErrorIs(t, err, ErrClosed)
	_, err = clone.GetNodeJSON("node")
	require.ErrorIs(t, err, ErrClosed)
	require.Nil(t, clone.Clone())
	require.Nil(t, clone.GetTasker())
	require.Nil(t, ctx.GetTasker())
	require.Nil(t, tasker.GetController())
	require.Nil(t, tasker.GetResource())
	require.Equal(t, before, nativeCalls.Load(), "invalidated views must not reach native code")
}
