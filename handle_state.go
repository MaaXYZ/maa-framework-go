package maa

import (
	"errors"
	"sync"
)

// ErrClosed reports an operation on a destroyed native object.
var ErrClosed = errors.New("maa: object is closed")

// ErrBorrowed reports an attempt to destroy a handle obtained from a getter or callback.
var ErrBorrowed = errors.New("maa: borrowed object cannot be destroyed")

// ErrBound reports an attempt to destroy a resource or controller still bound to a tasker.
var ErrBound = errors.New("maa: object is bound to a tasker")

// handleState keeps a native handle alive until calls already using it finish.
// Its cleanup runs once, after close and the last active call.
type handleState struct {
	mu       sync.Mutex
	handle   uintptr
	active   int
	bindings int
	closed   bool
	cleanup  func(uintptr)
}

func newHandleState(handle uintptr, cleanup func(uintptr)) *handleState {
	return &handleState{handle: handle, cleanup: cleanup}
}

func (s *handleState) begin() (uintptr, func(), error) {
	if s == nil {
		return 0, nil, ErrClosed
	}
	s.mu.Lock()
	if s.closed || s.handle == 0 {
		s.mu.Unlock()
		return 0, nil, ErrClosed
	}
	s.active++
	handle := s.handle
	s.mu.Unlock()
	return handle, s.end, nil
}

func (s *handleState) end() {
	s.mu.Lock()
	s.active--
	handle, cleanup := s.takeCleanupLocked()
	s.mu.Unlock()
	if cleanup != nil {
		cleanup(handle)
	}
}

func (s *handleState) takeCleanupLocked() (uintptr, func(uintptr)) {
	if !s.closed || s.active != 0 || s.cleanup == nil {
		return 0, nil
	}
	handle, cleanup := s.handle, s.cleanup
	s.handle = 0
	s.cleanup = nil
	return handle, cleanup
}

func (s *handleState) close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	if s.bindings != 0 {
		s.mu.Unlock()
		return ErrBound
	}
	s.closed = true
	handle, cleanup := s.takeCleanupLocked()
	s.mu.Unlock()
	if cleanup != nil {
		cleanup(handle)
	}
	return nil
}

func (s *handleState) addBinding() (uintptr, error) {
	if s == nil {
		return 0, ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.handle == 0 {
		return 0, ErrClosed
	}
	s.bindings++
	return s.handle, nil
}

func (s *handleState) removeBinding() {
	s.mu.Lock()
	s.bindings--
	s.mu.Unlock()
}

type taskerState struct {
	*handleState
	bindingsMu sync.Mutex
	resource   *Resource
	controller *Controller
}

var (
	taskerStates     sync.Map // map[uintptr]*taskerState
	resourceStates   sync.Map // map[uintptr]*handleState
	controllerStates sync.Map // map[uintptr]*handleState
)

func borrowTasker(handle uintptr) *Tasker {
	if state, ok := taskerStates.Load(handle); ok {
		return &Tasker{handle: handle, state: state.(*taskerState)}
	}
	return nil
}

func borrowResource(handle uintptr) *Resource {
	if state, ok := resourceStates.Load(handle); ok {
		return &Resource{handle: handle, state: state.(*handleState)}
	}
	return nil
}

func borrowController(handle uintptr) *Controller {
	if state, ok := controllerStates.Load(handle); ok {
		return &Controller{handle: handle, state: state.(*handleState)}
	}
	return nil
}
