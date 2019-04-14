// Based on the implementation from https://flaviocopes.com/golang-data-structure-queue/

package queue

import (
	"sync"

	"github.com/elyby/chrly/api/mojang"
)

type jobItem struct {
	Username  string
	RespondTo chan *mojang.SignedTexturesResponse
}

type jobsQueue struct {
	items []*jobItem
	lock  sync.RWMutex
}

func (s *jobsQueue) New() *jobsQueue {
	s.items = []*jobItem{}
	return s
}

func (s *jobsQueue) Enqueue(t *jobItem) {
	s.lock.Lock()
	s.items = append(s.items, t)
	s.lock.Unlock()
}

func (s *jobsQueue) Dequeue(n int) []*jobItem {
	s.lock.Lock()
	if n > s.Size() {
		n = s.Size()
	}

	items := s.items[0:n]
	s.items = s.items[n:len(s.items)]
	s.lock.Unlock()

	return items
}

func (s *jobsQueue) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *jobsQueue) Size() int {
	return len(s.items)
}
