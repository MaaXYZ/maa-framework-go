// Package store provides registries of per-handle callback-registration
// bookkeeping for the package's taskers, controllers, and resources.
package store

import "sync"

// Store is a registry of values keyed by native handle.
//
// Set, Get, and Del take no lock: the caller must hold Lock for the duration
// of any sequence of those calls. Update acquires the lock itself and must
// not be called while the caller holds Lock.
type Store[T any] struct {
	data map[uintptr]T
	mu   sync.RWMutex
}

// New returns an empty Store ready for use.
func New[T any]() *Store[T] {
	return &Store[T]{data: make(map[uintptr]T)}
}

// Lock locks the store exclusively, guarding the Set, Get, and Del calls made
// while it is held.
func (s *Store[T]) Lock() {
	s.mu.Lock()
}

// Unlock releases the store. Every Lock call must be paired with exactly one
// Unlock.
func (s *Store[T]) Unlock() {
	s.mu.Unlock()
}

// Set stores value under handle, replacing any previous value. The caller
// must hold Lock.
func (s *Store[T]) Set(handle uintptr, value T) {
	s.data[handle] = value
}

// Get returns the value stored under handle, or the zero value of T when
// handle is absent. The caller must hold Lock.
func (s *Store[T]) Get(handle uintptr) T {
	return s.data[handle]
}

// Del removes the value stored under handle. Deleting an absent handle is a
// no-op. The caller must hold Lock.
func (s *Store[T]) Del(handle uintptr) {
	delete(s.data, handle)
}

// Update locks the store, passes a pointer to a copy of the value stored
// under handle — the zero value of T when absent — to fn, stores the modified
// copy, and unlocks. Unlike Set, Get, and Del, Update must be called without
// holding Lock.
func (s *Store[T]) Update(handle uintptr, fn func(*T)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value := s.data[handle]
	fn(&value)
	s.data[handle] = value
}

// TaskerStoreValue records the callback registrations for one tasker: the
// event callbacks backing its sinks and the event callbacks backing its
// context-scoped sinks, each keyed by sink ID.
type TaskerStoreValue struct {
	SinkIDToEventCallbackID        map[int64]uint64
	ContextSinkIDToEventCallbackID map[int64]uint64
}

// CtrlStoreValue records the callback registrations for one controller: the
// event callbacks backing its sinks, keyed by sink ID, and the shared
// custom-controller callback registration.
type CtrlStoreValue struct {
	SinkIDToEventCallbackID     map[int64]uint64
	CustomControllerCallbacksID uint64
}

// ResStoreValue records the callback registrations for one resource: the
// event callbacks backing its sinks, keyed by sink ID, and the custom
// recognizer and custom action callback registrations, keyed by name.
type ResStoreValue struct {
	SinkIDToEventCallbackID     map[int64]uint64
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

// TaskerStore, CtrlStore, and ResStore track the callback registrations of
// the package's taskers, controllers, and resources.
var (
	TaskerStore = New[TaskerStoreValue]()
	CtrlStore   = New[CtrlStoreValue]()
	ResStore    = New[ResStoreValue]()
)
