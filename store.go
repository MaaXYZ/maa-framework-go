package maa

import "unsafe"

type store[T any] struct {
	data map[uintptr]T
}

func newStore[T any]() *store[T] {
	return &store[T]{data: make(map[uintptr]T)}
}

func (s *store[T]) set(handle unsafe.Pointer, value T) {
	key := uintptr(handle)
	s.data[key] = value
}

func (s *store[T]) get(handle unsafe.Pointer) T {
	key := uintptr(handle)
	return s.data[key]
}

func (s *store[T]) del(handle unsafe.Pointer) {
	key := uintptr(handle)
	delete(s.data, key)
}
