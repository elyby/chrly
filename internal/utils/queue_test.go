package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Run("Enqueue", func(t *testing.T) {
		s := NewQueue[string]()
		require.Equal(t, 1, s.Enqueue("username1"))
		require.Equal(t, 2, s.Enqueue("username2"))
		require.Equal(t, 3, s.Enqueue("username3"))
	})

	t.Run("Dequeue", func(t *testing.T) {
		s := NewQueue[string]()
		s.Enqueue("username1")
		s.Enqueue("username2")
		s.Enqueue("username3")
		s.Enqueue("username4")
		s.Enqueue("username5")

		items, queueLen := s.Dequeue(2)
		require.Len(t, items, 2)
		require.Equal(t, 3, queueLen)
		require.Equal(t, "username1", items[0])
		require.Equal(t, "username2", items[1])

		items, queueLen = s.Dequeue(40)
		require.Len(t, items, 3)
		require.Equal(t, 0, queueLen)
		require.Equal(t, "username3", items[0])
		require.Equal(t, "username4", items[1])
		require.Equal(t, "username5", items[2])
	})

	t.Run("Len", func(t *testing.T) {
		s := NewQueue[string]()
		s.Enqueue("username1")
		s.Enqueue("username2")
		s.Enqueue("username3")

		require.Equal(t, 3, s.Len())
	})
}
