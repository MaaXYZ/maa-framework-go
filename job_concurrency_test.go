package maa

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// jobConcurrencyTimeout only turns a deadlocked handshake into a test failure.
// No ordering in this file depends on elapsed time: every step is released by a
// channel handshake.
const jobConcurrencyTimeout = 10 * time.Second

// jobConcurrencyAwait receives one handshake value. It must be called from the
// goroutine that owns t, never from a worker goroutine.
func jobConcurrencyAwait[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(jobConcurrencyTimeout):
		var zero T
		t.Fatal("timed out waiting for a concurrency handshake")
		return zero
	}
}

// jobConcurrencyCall runs a status query on a separate goroutine so a query
// that wrongly blocks behind Wait fails the test instead of hanging it.
func jobConcurrencyCall[T any](t *testing.T, fn func() T) T {
	t.Helper()
	result := make(chan T, 1)
	go func() { result <- fn() }()
	return jobConcurrencyAwait(t, result)
}

// jobConcurrencyRelease returns an idempotent closer so every failure path can
// unblock a waiting worker.
func jobConcurrencyRelease(ch chan struct{}) func() {
	var once sync.Once
	return func() { once.Do(func() { close(ch) }) }
}

func TestJob_ConcurrentWaitAndStatus(t *testing.T) {
	t.Run("StatusReadableWhileWaitBlocked", func(t *testing.T) {
		waitEntered := make(chan struct{})
		releaseWait := make(chan struct{})
		release := jobConcurrencyRelease(releaseWait)
		defer release()

		var waitCalls atomic.Int64
		job := newJob(1,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				close(waitEntered)
				<-releaseWait
				return StatusSuccess
			},
		)

		waitDone := make(chan struct{})
		go func() {
			defer close(waitDone)
			job.Wait()
		}()

		jobConcurrencyAwait(t, waitEntered)

		// Status and the predicate helpers must stay callable while Wait is
		// blocked inside waitFunc.
		require.Equal(t, StatusRunning, jobConcurrencyCall(t, job.Status))
		require.True(t, jobConcurrencyCall(t, job.Running))
		require.False(t, jobConcurrencyCall(t, job.Success))
		require.False(t, jobConcurrencyCall(t, job.Done))

		select {
		case <-waitDone:
			t.Fatal("Wait returned before waitFunc was released")
		default:
		}

		// These readers overlap the terminal store. Their fixed read count keeps
		// the overlap independent of scheduling.
		const (
			readers        = 4
			readsPerReader = 64
		)
		readerGo := make(chan struct{})
		releaseReaders := jobConcurrencyRelease(readerGo)
		defer releaseReaders()
		readerStarted := make(chan struct{}, readers)
		readerDone := make(chan struct{}, readers)
		for range readers {
			go func() {
				readerStarted <- struct{}{}
				<-readerGo
				for range readsPerReader {
					_ = job.Status()
					_ = job.Success()
					_ = job.Done()
				}
				readerDone <- struct{}{}
			}()
		}
		for range readers {
			jobConcurrencyAwait(t, readerStarted)
		}
		releaseReaders()
		release()
		jobConcurrencyAwait(t, waitDone)
		for range readers {
			jobConcurrencyAwait(t, readerDone)
		}

		require.EqualValues(t, 1, waitCalls.Load(), "a successful wait runs waitFunc once")
		require.Equal(t, StatusSuccess, job.Status())
		require.True(t, job.Success())
		require.True(t, job.Done())
	})

	t.Run("AllCallersFinishAndWaitFuncRunsOnce", func(t *testing.T) {
		const waiters = 4
		waitEntered := make(chan struct{}, waiters)
		releaseWait := make(chan struct{})
		release := jobConcurrencyRelease(releaseWait)
		defer release()

		var waitCalls atomic.Int64
		job := newJob(2,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				waitEntered <- struct{}{}
				<-releaseWait
				return StatusSuccess
			},
		)

		start := make(chan struct{})
		results := make(chan Status, waiters)
		var ready sync.WaitGroup
		ready.Add(waiters)
		for range waiters {
			go func() {
				ready.Done()
				<-start
				job.Wait()
				results <- job.Status()
			}()
		}
		ready.Wait()
		close(start)

		// The first caller blocks inside waitFunc. Any extra entry, which only
		// an unsynchronized Wait can produce, is drained before the release.
		jobConcurrencyAwait(t, waitEntered)
		for drain := true; drain; {
			select {
			case <-waitEntered:
			default:
				drain = false
			}
		}

		release()
		for range waiters {
			require.Equal(t, StatusSuccess, jobConcurrencyAwait(t, results))
		}
		require.EqualValues(t, 1, waitCalls.Load(), "serialized waiters share one waitFunc result")
	})

	t.Run("TerminalStatusAfterOwnerClose", func(t *testing.T) {
		state := newHandleState(1, func(uintptr) {})
		t.Cleanup(func() { _ = state.close() })
		state.jobStatus = func(uintptr, int64) Status { return StatusSuccess }

		var waitCalls atomic.Int64
		job := newJob(3,
			func(int64) Status { return StatusSuccess },
			func(int64) Status {
				waitCalls.Add(1)
				return StatusSuccess
			},
			state,
		)

		job.Wait()
		require.Equal(t, StatusSuccess, job.Status())
		require.EqualValues(t, 1, waitCalls.Load())

		require.NoError(t, state.close())
		require.ErrorIs(t, job.Error(), ErrClosed)

		const readers = 8
		results := make(chan Status, readers)
		var readersDone sync.WaitGroup
		readersDone.Add(readers)
		for range readers {
			go func() {
				defer readersDone.Done()
				results <- job.Status()
				job.Wait()
			}()
		}
		readersDone.Wait()

		for range readers {
			require.Equal(t, StatusSuccess, jobConcurrencyAwait(t, results))
		}
		require.EqualValues(t, 1, waitCalls.Load(), "Wait after owner close must not run waitFunc")
		require.True(t, job.Success())
		require.True(t, job.Done())
	})
}

func TestTaskJob_ConcurrentWaitAndStatus(t *testing.T) {
	t.Run("StatusReadableWhileWaitBlocked", func(t *testing.T) {
		waitEntered := make(chan struct{})
		releaseWait := make(chan struct{})
		release := jobConcurrencyRelease(releaseWait)
		defer release()

		var waitCalls atomic.Int64
		job := newTaskJob(1,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				close(waitEntered)
				<-releaseWait
				return StatusSuccess
			},
			nil, nil, nil,
		)

		waitDone := make(chan struct{})
		go func() {
			defer close(waitDone)
			job.Wait()
		}()

		jobConcurrencyAwait(t, waitEntered)

		require.Equal(t, StatusRunning, jobConcurrencyCall(t, job.Status))
		require.True(t, jobConcurrencyCall(t, job.Running))
		require.False(t, jobConcurrencyCall(t, job.Success))
		require.False(t, jobConcurrencyCall(t, job.Done))

		select {
		case <-waitDone:
			t.Fatal("Wait returned before waitFunc was released")
		default:
		}

		const (
			readers        = 4
			readsPerReader = 64
		)
		readerGo := make(chan struct{})
		releaseReaders := jobConcurrencyRelease(readerGo)
		defer releaseReaders()
		readerStarted := make(chan struct{}, readers)
		readerDone := make(chan struct{}, readers)
		for range readers {
			go func() {
				readerStarted <- struct{}{}
				<-readerGo
				for range readsPerReader {
					_ = job.Status()
					_ = job.Success()
					_ = job.Done()
				}
				readerDone <- struct{}{}
			}()
		}
		for range readers {
			jobConcurrencyAwait(t, readerStarted)
		}
		releaseReaders()
		release()
		jobConcurrencyAwait(t, waitDone)
		for range readers {
			jobConcurrencyAwait(t, readerDone)
		}

		require.EqualValues(t, 1, waitCalls.Load(), "a successful wait runs waitFunc once")
		require.Equal(t, StatusSuccess, job.Status())
		require.True(t, job.Success())
		require.True(t, job.Done())
	})

	t.Run("AllCallersFinishAndWaitFuncRunsOnce", func(t *testing.T) {
		const waiters = 4
		waitEntered := make(chan struct{}, waiters)
		releaseWait := make(chan struct{})
		release := jobConcurrencyRelease(releaseWait)
		defer release()

		var waitCalls atomic.Int64
		job := newTaskJob(2,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				waitEntered <- struct{}{}
				<-releaseWait
				return StatusSuccess
			},
			nil, nil, nil,
		)

		start := make(chan struct{})
		results := make(chan Status, waiters)
		var ready sync.WaitGroup
		ready.Add(waiters)
		for range waiters {
			go func() {
				ready.Done()
				<-start
				job.Wait()
				results <- job.Status()
			}()
		}
		ready.Wait()
		close(start)

		jobConcurrencyAwait(t, waitEntered)
		for drain := true; drain; {
			select {
			case <-waitEntered:
			default:
				drain = false
			}
		}

		release()
		for range waiters {
			require.Equal(t, StatusSuccess, jobConcurrencyAwait(t, results))
		}
		require.EqualValues(t, 1, waitCalls.Load(), "serialized waiters share one waitFunc result")
	})

	t.Run("TerminalStatusAfterOwnerClose", func(t *testing.T) {
		state := newHandleState(1, func(uintptr) {})
		t.Cleanup(func() { _ = state.close() })
		state.jobStatus = func(uintptr, int64) Status { return StatusSuccess }

		var waitCalls atomic.Int64
		job := newTaskJob(3,
			func(int64) Status { return StatusSuccess },
			func(int64) Status {
				waitCalls.Add(1)
				return StatusSuccess
			},
			nil, nil, nil,
			state,
		)

		job.Wait()
		require.Equal(t, StatusSuccess, job.Status())
		require.EqualValues(t, 1, waitCalls.Load())

		require.NoError(t, state.close())
		require.ErrorIs(t, job.Error(), ErrClosed)

		const readers = 8
		results := make(chan Status, readers)
		var readersDone sync.WaitGroup
		readersDone.Add(readers)
		for range readers {
			go func() {
				defer readersDone.Done()
				results <- job.Status()
				job.Wait()
			}()
		}
		readersDone.Wait()

		for range readers {
			require.Equal(t, StatusSuccess, jobConcurrencyAwait(t, results))
		}
		require.EqualValues(t, 1, waitCalls.Load(), "Wait after owner close must not run waitFunc")
		require.True(t, job.Success())
		require.True(t, job.Done())
	})
}

func TestJob_WaitRetriesInvalid(t *testing.T) {
	t.Run("SequentialRetry", func(t *testing.T) {
		var waitCalls atomic.Int64
		job := newJob(1,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				if waitCalls.Add(1) == 1 {
					return StatusInvalid
				}
				return StatusSuccess
			},
		)

		job.Wait()
		require.EqualValues(t, 1, waitCalls.Load())
		require.Equal(t, StatusRunning, job.Status(), "an invalid wait result must leave the job waitable")

		job.Wait()
		require.EqualValues(t, 2, waitCalls.Load())
		require.Equal(t, StatusSuccess, job.Status())
	})

	t.Run("ConcurrentRetry", func(t *testing.T) {
		const waiters = 2
		var waitCalls atomic.Int64
		job := newJob(2,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				if waitCalls.Add(1) == 1 {
					return StatusInvalid
				}
				return StatusSuccess
			},
		)

		start := make(chan struct{})
		waitDone := make(chan struct{}, waiters)
		var ready sync.WaitGroup
		ready.Add(waiters)
		for range waiters {
			go func() {
				ready.Done()
				<-start
				job.Wait()
				waitDone <- struct{}{}
			}()
		}
		ready.Wait()
		close(start)

		for range waiters {
			jobConcurrencyAwait(t, waitDone)
		}
		require.EqualValues(t, 2, waitCalls.Load(), "the first invalid wait must be retried")
		require.Equal(t, StatusSuccess, job.Status())
	})
}

func TestJob_FailedConcurrent(t *testing.T) {
	t.Run("JobSkipsWait", func(t *testing.T) {
		expectedErr := errors.New("job submission failed")
		job := newFailedJob(expectedErr)

		const callers = 8
		results := make(chan Status, callers)
		var callersDone sync.WaitGroup
		callersDone.Add(callers)
		for range callers {
			go func() {
				defer callersDone.Done()
				job.Wait()
				results <- job.Status()
			}()
		}
		callersDone.Wait()

		require.ErrorIs(t, job.Error(), expectedErr)
		for range callers {
			require.Equal(t, StatusFailure, jobConcurrencyAwait(t, results))
		}
		require.True(t, job.Failure())
		require.True(t, job.Done())
	})

	t.Run("TaskJobSkipsWait", func(t *testing.T) {
		expectedErr := errors.New("job submission failed")
		var waitCalls atomic.Int64
		job := newTaskJob(1,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				return StatusSuccess
			},
			nil, nil, expectedErr,
		)

		const callers = 8
		var callersDone sync.WaitGroup
		callersDone.Add(callers)
		for range callers {
			go func() {
				defer callersDone.Done()
				job.Wait()
			}()
		}
		callersDone.Wait()

		require.EqualValues(t, 0, waitCalls.Load(), "a failed TaskJob must not wait")
		require.Equal(t, StatusFailure, job.Status())
		require.True(t, job.Failure())
		require.True(t, job.Done())
	})

	t.Run("ClosedOwnerSkipsWait", func(t *testing.T) {
		state := newHandleState(1, func(uintptr) {})
		t.Cleanup(func() { _ = state.close() })
		require.NoError(t, state.close())

		var waitCalls atomic.Int64
		job := newJob(1,
			func(int64) Status { return StatusRunning },
			func(int64) Status {
				waitCalls.Add(1)
				return StatusSuccess
			},
			state,
		)

		const callers = 8
		results := make(chan Status, callers)
		var callersDone sync.WaitGroup
		callersDone.Add(callers)
		for range callers {
			go func() {
				defer callersDone.Done()
				job.Wait()
				results <- job.Status()
			}()
		}
		callersDone.Wait()

		require.ErrorIs(t, job.Error(), ErrClosed)
		require.EqualValues(t, 0, waitCalls.Load(), "a job on a closed owner must not wait")
		for range callers {
			require.Equal(t, StatusFailure, jobConcurrencyAwait(t, results))
		}
	})
}
