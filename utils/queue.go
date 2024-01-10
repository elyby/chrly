package utils

import (
	"sync"
)

type Queue[T any] struct {
	lock  sync.Mutex
	items []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: []T{},
	}
}

func (s *Queue[T]) Enqueue(item T) int {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, item)

	return len(s.items)
}

func (s *Queue[T]) Dequeue(n int) ([]T, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.items)
	if n > l {
		n = l
	}

	items := s.items[0:n]
	s.items = s.items[n:l]

	return items, l - n
}
