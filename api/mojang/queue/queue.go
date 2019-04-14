// Based on the implementation from https://flaviocopes.com/golang-data-structure-queue/

package queue

import (
	"sync"

	"github.com/elyby/chrly/api/mojang"
)

type Job struct {
	Username  string
	RespondTo chan *mojang.SignedTexturesResponse
}

type JobsQueue struct {
	items []*Job
	lock  sync.RWMutex
}

func (s *JobsQueue) New() *JobsQueue {
	s.items = []*Job{}
	return s
}

func (s *JobsQueue) Enqueue(t *Job) {
	s.lock.Lock()
	s.items = append(s.items, t)
	s.lock.Unlock()
}

func (s *JobsQueue) Dequeue(n int) []*Job {
	s.lock.Lock()
	if n > s.Size() {
		n = s.Size()
	}

	items := s.items[0:n]
	s.items = s.items[n:len(s.items)]
	s.lock.Unlock()

	return items
}

func (s *JobsQueue) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *JobsQueue) Size() int {
	return len(s.items)
}
