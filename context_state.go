package maa

import "sync"

// contextState bounds all copies and clones of a callback context to its call.
type contextState struct {
	mu     sync.Mutex
	cond   *sync.Cond
	active int
	closed bool
}

func newCallbackContext(handle uintptr) *Context {
	state := &contextState{}
	state.cond = sync.NewCond(&state.mu)
	return &Context{handle: handle, state: state}
}

func (s *contextState) begin() (func(), error) {
	if s == nil {
		return nil, ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, ErrClosed
	}
	s.active++
	return s.end, nil
}

func (s *contextState) end() {
	s.mu.Lock()
	s.active--
	if s.active == 0 {
		s.cond.Broadcast()
	}
	s.mu.Unlock()
}

func (s *contextState) invalidate() {
	s.mu.Lock()
	s.closed = true
	for s.active != 0 {
		s.cond.Wait()
	}
	s.mu.Unlock()
}

func (ctx *Context) invalidate() {
	if ctx != nil && ctx.state != nil {
		ctx.state.invalidate()
	}
}
