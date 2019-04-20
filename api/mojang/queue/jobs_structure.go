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
	lock  sync.Mutex
	items []*jobItem
}

func (s *jobsQueue) New() *jobsQueue {
	s.items = []*jobItem{}
	return s
}

func (s *jobsQueue) Enqueue(t *jobItem) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = append(s.items, t)
}

func (s *jobsQueue) Dequeue(n int) []*jobItem {
	s.lock.Lock()
	defer s.lock.Unlock()

	if n > s.Size() {
		n = s.Size()
	}

	items := s.items[0:n]
	s.items = s.items[n:len(s.items)]

	return items
}

func (s *jobsQueue) IsEmpty() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.items) == 0
}

func (s *jobsQueue) Size() int {
	return len(s.items)
}
