package store

import "unsafe"

type Store[T any] struct {
	data map[uintptr]T
}

func New[T any]() *Store[T] {
	return &Store[T]{data: make(map[uintptr]T)}
}

func (s *Store[T]) Set(handle unsafe.Pointer, value T) {
	key := uintptr(handle)
	s.data[key] = value
}

func (s *Store[T]) Get(handle unsafe.Pointer) T {
	key := uintptr(handle)
	return s.data[key]
}

func (s *Store[T]) Del(handle unsafe.Pointer) {
	key := uintptr(handle)
	delete(s.data, key)
}
