package store

type Store[T any] struct {
	data map[uintptr]T
}

func New[T any]() *Store[T] {
	return &Store[T]{data: make(map[uintptr]T)}
}

func (s *Store[T]) Set(handle uintptr, value T) {
	key := handle
	s.data[key] = value
}

func (s *Store[T]) Get(handle uintptr) T {
	key := handle
	return s.data[key]
}

func (s *Store[T]) Del(handle uintptr) {
	key := handle
	delete(s.data, key)
}
