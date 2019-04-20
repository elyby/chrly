package queue

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	assert := testify.New(t)

	s := createQueue()
	s.Enqueue(&jobItem{Username: "username1"})
	s.Enqueue(&jobItem{Username: "username2"})
	s.Enqueue(&jobItem{Username: "username3"})

	assert.Equal(3, s.Size())
}

func TestDequeueN(t *testing.T) {
	assert := testify.New(t)

	s := createQueue()
	s.Enqueue(&jobItem{Username: "username1"})
	s.Enqueue(&jobItem{Username: "username2"})
	s.Enqueue(&jobItem{Username: "username3"})
	s.Enqueue(&jobItem{Username: "username4"})

	items := s.Dequeue(2)
	assert.Len(items, 2)
	assert.Equal("username1", items[0].Username)
	assert.Equal("username2", items[1].Username)
	assert.Equal(2, s.Size())

	items = s.Dequeue(40)
	assert.Len(items, 2)
	assert.Equal("username3", items[0].Username)
	assert.Equal("username4", items[1].Username)
	assert.True(s.IsEmpty())
}

func createQueue() *jobsQueue {
	queue := &jobsQueue{}
	queue.New()

	return queue
}
