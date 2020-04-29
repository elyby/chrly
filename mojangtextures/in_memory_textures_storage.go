package mojangtextures

import (
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"
)

type inMemoryItem struct {
	textures  *mojang.SignedTexturesResponse
	timestamp int64
}

type InMemoryTexturesStorage struct {
	GCPeriod time.Duration
	Duration time.Duration

	once sync.Once
	lock sync.RWMutex
	data map[string]*inMemoryItem
	done chan struct{}
}

func NewInMemoryTexturesStorage() *InMemoryTexturesStorage {
	storage := &InMemoryTexturesStorage{
		GCPeriod: 10 * time.Second,
		Duration: time.Minute + 10*time.Second,
		data:     make(map[string]*inMemoryItem),
	}

	return storage
}

func (s *InMemoryTexturesStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	item, exists := s.data[uuid]
	validRange := s.getMinimalNotExpiredTimestamp()
	if !exists || validRange > item.timestamp {
		return nil, nil
	}

	return item.textures, nil
}

func (s *InMemoryTexturesStorage) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	s.once.Do(s.start)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[uuid] = &inMemoryItem{
		textures:  textures,
		timestamp: unixNanoToUnixMicro(time.Now().UnixNano()),
	}
}

func (s *InMemoryTexturesStorage) start() {
	s.done = make(chan struct{})
	ticker := time.NewTicker(s.GCPeriod)
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.gc()
			}
		}
	}()
}

func (s *InMemoryTexturesStorage) Stop() {
	close(s.done)
}

func (s *InMemoryTexturesStorage) gc() {
	s.lock.Lock()
	defer s.lock.Unlock()

	maxTime := s.getMinimalNotExpiredTimestamp()
	for uuid, value := range s.data {
		if maxTime > value.timestamp {
			delete(s.data, uuid)
		}
	}
}

func (s *InMemoryTexturesStorage) getMinimalNotExpiredTimestamp() int64 {
	return unixNanoToUnixMicro(time.Now().Add(s.Duration * time.Duration(-1)).UnixNano())
}

func unixNanoToUnixMicro(unixNano int64) int64 {
	return unixNano / 10e5
}
