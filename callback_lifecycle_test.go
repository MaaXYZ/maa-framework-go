package maa

import (
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/native"
	"github.com/stretchr/testify/require"
)

func TestHandleState_InvalidJobNeedsIdleProof(t *testing.T) {
	destroyed, idle := false, false
	state := newHandleState(1, func(uintptr) { destroyed = true })
	state.jobStatus = func(uintptr, int64) Status { return StatusInvalid }
	state.idleProbe = func(uintptr) bool { return idle }
	job := newJob(1, func(int64) Status { return StatusInvalid }, func(int64) Status { return StatusInvalid }, state)
	job.Wait()
	require.Equal(t, StatusInvalid, job.Status())
	require.ErrorIs(t, state.close(), ErrInUse)
	require.False(t, destroyed)
	idle = true
	require.NoError(t, state.close())
	require.True(t, destroyed)
}

func TestController_InvalidJobCloseUsesBarrier(t *testing.T) {
	const handle = uintptr(876543)
	initControllerStore(handle)
	ctrl := newOwnedController(handle)
	var destroyed, posts atomic.Int32
	var accepting atomic.Bool
	var barrierStatus atomic.Int32
	barrierStatus.Store(int32(StatusRunning))
	replaceNativeForTest(t, &native.MaaControllerDestroy, func(uintptr) { destroyed.Add(1) })
	replaceNativeForTest(t, &native.MaaControllerStatus, func(_ uintptr, id int64) int32 {
		if id != 1 {
			return barrierStatus.Load()
		}
		return int32(StatusInvalid)
	})
	replaceNativeForTest(t, &native.MaaControllerPostInactive, func(uintptr) int64 {
		posts.Add(1)
		if accepting.Load() {
			return int64(posts.Load()) + 1
		}
		return 0
	})
	job := newJob(1, ctrl.status, ctrl.wait, ctrl.state)
	require.Equal(t, StatusInvalid, job.Status())
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.Zero(t, destroyed.Load())
	accepting.Store(true)
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.EqualValues(t, 2, posts.Load())
	barrierStatus.Store(int32(StatusInvalid))
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.EqualValues(t, 3, posts.Load(), "an invalidated barrier must be reposted")
	barrierStatus.Store(int32(StatusSuccess))
	require.NoError(t, ctrl.Destroy())
	require.NoError(t, ctrl.Destroy())
	require.EqualValues(t, 1, destroyed.Load())
}

func TestController_StopInvalidatesCloseBarrier(t *testing.T) {
	const handle = uintptr(876544)
	initControllerStore(handle)
	ctrl := newOwnedController(handle)
	var posts atomic.Int32
	replaceNativeForTest(t, &native.MaaControllerDestroy, func(uintptr) {})
	replaceNativeForTest(t, &native.MaaControllerStatus, func(uintptr, int64) int32 { return int32(StatusSuccess) })
	replaceNativeForTest(t, &native.MaaControllerPostInactive, func(uintptr) int64 { return int64(posts.Add(1)) })
	ctrl.state.markJobsUncertain()
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	// A subsequent stop invalidates the first completed barrier's proof.
	ctrl.state.markJobsUncertain()
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.EqualValues(t, 2, posts.Load())
	require.NoError(t, ctrl.Destroy())
}

func TestResource_InvalidJobCloseChecksLoader(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	var idle atomic.Bool
	replaceNativeForTest(t, &native.MaaResourceStatus, func(uintptr, int64) int32 { return int32(StatusInvalid) })
	replaceNativeForTest(t, &native.MaaResourceClear, func(uintptr) bool { return idle.Load() })
	job := newJob(1, res.status, res.wait, res.state)
	require.Equal(t, StatusInvalid, job.Status())
	require.ErrorIs(t, res.Destroy(), ErrInUse)
	idle.Store(true)
	require.NoError(t, res.Destroy())
}

func TestTasker_StopMarksBoundOwnersUncertain(t *testing.T) {
	res, err := NewResource()
	require.NoError(t, err)
	ctrl, err := NewBlankController()
	require.NoError(t, err)
	tasker, err := NewTasker()
	require.NoError(t, err)
	require.NoError(t, tasker.BindResource(res))
	require.NoError(t, tasker.BindController(ctrl))
	replaceNativeForTest(t, &native.MaaTaskerPostStop, func(uintptr) int64 { return 42 })
	replaceNativeForTest(t, &native.MaaTaskerStatus, func(uintptr, int64) int32 { return int32(StatusSuccess) })
	_, err = tasker.PostStop()
	require.NoError(t, err)
	for _, state := range []*handleState{res.state, ctrl.state} {
		state.mu.Lock()
		uncertain := state.jobsUncertain
		state.mu.Unlock()
		require.True(t, uncertain)
	}
	require.NoError(t, tasker.Destroy())
	require.NoError(t, res.Destroy())
	// The real controller executes the close barrier asynchronously.
	require.ErrorIs(t, ctrl.Destroy(), ErrInUse)
	require.Eventually(t, func() bool { return ctrl.Destroy() == nil }, jobConcurrencyTimeout, jobReapInterval/10)
}

type cleanupCallbackController struct {
	BlankController
	keyUp func() bool
	click func() bool
}

func (c *cleanupCallbackController) KeyUp(int32) bool        { return c.keyUp() }
func (c *cleanupCallbackController) Click(int32, int32) bool { return c.click() }

func TestCustomController_CallbackLifetimeAndCleanup(t *testing.T) {
	var id uintptr
	replaceNativeForTest(t, &native.MaaCustomControllerCreate, func(_ unsafe.Pointer, arg uintptr) uintptr { id = arg; return 876545 })
	var owner *Controller
	var cleanupCalls atomic.Int32
	var destroyErr error
	entered, release := make(chan struct{}), make(chan struct{})
	unblock := jobConcurrencyRelease(release)
	defer unblock()
	impl := &cleanupCallbackController{
		keyUp: func() bool {
			cleanupCalls.Add(1)
			destroyErr = owner.Destroy()
			return true
		},
		click: func() bool { close(entered); <-release; return true },
	}
	replaceNativeForTest(t, &native.MaaControllerDestroy, func(uintptr) { _KeyUp(1, id) })
	var err error
	owner, err = NewCustomController(impl)
	require.NoError(t, err)
	called := make(chan uintptr, 1)
	go func() { called <- _ClickAgent(0, 0, id) }()
	jobConcurrencyAwait(t, entered)
	require.ErrorIs(t, owner.Destroy(), ErrInCallback)
	unblock()
	require.EqualValues(t, 1, jobConcurrencyAwait(t, called))
	require.NoError(t, owner.Destroy())
	require.ErrorIs(t, destroyErr, ErrInCallback)
	require.EqualValues(t, 1, cleanupCalls.Load())
	_KeyUp(1, id)
	require.EqualValues(t, 1, cleanupCalls.Load(), "no callbacks after successful destruction")
	require.NoError(t, owner.Destroy())
}

func TestTasker_ConcurrentBindingGetters(t *testing.T) {
	tasker, err := NewTasker()
	require.NoError(t, err)
	resources := make([]*Resource, 2)
	controllers := make([]*Controller, 2)
	for i := range 2 {
		resources[i], err = NewResource()
		require.NoError(t, err)
		controllers[i], err = NewBlankController()
		require.NoError(t, err)
	}
	require.NoError(t, tasker.BindResource(resources[0]))
	require.NoError(t, tasker.BindController(controllers[0]))
	stop := make(chan struct{})
	unblock := jobConcurrencyRelease(stop)
	defer unblock()
	done := make(chan bool, 1)
	ready := make(chan struct{})
	go func() {
		close(ready)
		ok := true
		for {
			select {
			case <-stop:
				done <- ok
				return
			default:
				res, ctrl := tasker.GetResource(), tasker.GetController()
				ok = ok && res != nil && ctrl != nil
				if res != nil {
					ok = ok && (res.state == resources[0].state || res.state == resources[1].state) && !res.owned
				}
				if ctrl != nil {
					ok = ok && (ctrl.state == controllers[0].state || ctrl.state == controllers[1].state) && !ctrl.owned
				}
			}
		}
	}()
	jobConcurrencyAwait(t, ready)
	for i := range 100 {
		require.NoError(t, tasker.BindResource(resources[i%2]))
		require.NoError(t, tasker.BindController(controllers[i%2]))
	}
	unblock()
	require.True(t, jobConcurrencyAwait(t, done))
	require.NoError(t, tasker.Destroy())
	for i := range 2 {
		require.NoError(t, resources[i].Destroy())
		require.NoError(t, controllers[i].Destroy())
	}
}

func TestTasker_StopDuringCustomControllerCallback(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	unblock := jobConcurrencyRelease(release)
	defer unblock()
	ctrl, err := NewCustomController(&cleanupCallbackController{
		click: func() bool { close(entered); <-release; return true },
	})
	require.NoError(t, err)
	res, err := NewResource()
	require.NoError(t, err)
	tasker, err := NewTasker()
	require.NoError(t, err)
	t.Cleanup(func() {
		unblock()
		require.Eventually(t, func() bool { return tasker.Destroy() == nil }, jobConcurrencyTimeout, jobReapInterval/10)
		require.Eventually(t, func() bool { return ctrl.Destroy() == nil }, jobConcurrencyTimeout, jobReapInterval/10)
		require.NoError(t, res.Destroy())
	})
	connect, err := ctrl.PostConnect()
	require.NoError(t, err)
	require.True(t, connect.Wait().Success())
	require.NoError(t, tasker.BindResource(res))
	require.NoError(t, tasker.BindController(ctrl))
	click, err := ctrl.PostClick(0, 0)
	require.NoError(t, err)
	jobConcurrencyAwait(t, entered)
	stop, err := tasker.PostStop()
	require.NoError(t, err)
	require.True(t, stop.Wait().Success())
	// Stop invalidates the action ID before the blocked callback returns.
	require.Equal(t, StatusInvalid, click.Wait().Status())
	require.Eventually(t, func() bool { return tasker.Destroy() == nil }, jobConcurrencyTimeout, jobReapInterval/10)
	require.ErrorIs(t, ctrl.Destroy(), ErrInCallback)
	unblock()
	// The custom call may finish before the native process publishes completion.
	require.Eventually(t, func() bool { return ctrl.Destroy() == nil }, jobConcurrencyTimeout, jobReapInterval/10)
	require.NoError(t, res.Destroy())
}
