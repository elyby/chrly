package mojangtextures

import (
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/raven-go"

	"github.com/elyby/chrly/api/mojang"

	"github.com/tevino/abool"
)

var now = time.Now

type inMemoryItem struct {
	textures  *mojang.SignedTexturesResponse
	timestamp int64
}

type InMemoryTexturesStorage struct {
	GCPeriod time.Duration
	Duration time.Duration

	lock    sync.RWMutex
	data    map[string]*inMemoryItem
	working *abool.AtomicBool
}

func NewInMemoryTexturesStorage() *InMemoryTexturesStorage {
	storage := &InMemoryTexturesStorage{
		GCPeriod: 10 * time.Second,
		Duration: time.Minute + 10*time.Second,
		data:     make(map[string]*inMemoryItem),
	}

	return storage
}

func (s *InMemoryTexturesStorage) Start() {
	if s.working == nil {
		s.working = abool.New()
	}

	if !s.working.IsSet() {
		go func() {
			time.Sleep(s.GCPeriod)
			// TODO: this can be reimplemented in future with channels, but right now I have no idea how to make it right
			for s.working.IsSet() {
				start := time.Now()
				s.gc()
				time.Sleep(s.GCPeriod - time.Since(start))
			}
		}()
	}

	s.working.Set()
}

func (s *InMemoryTexturesStorage) Stop() {
	s.working.UnSet()
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
	var timestamp int64
	if textures != nil {
		decoded, err := textures.DecodeTextures()
		if err != nil {
			tags := map[string]string{
				"textures.id":   textures.Id,
				"textures.name": textures.Name,
			}

			for i, prop := range textures.Props {
				tags[fmt.Sprintf("textures.props[%d].name", i)] = prop.Name
				tags[fmt.Sprintf("textures.props[%d].value", i)] = prop.Value
				tags[fmt.Sprintf("textures.props[%d].signature", i)] = prop.Signature
			}

			raven.CaptureErrorAndWait(err, tags)

			panic(err)
		}

		timestamp = decoded.Timestamp
	} else {
		timestamp = unixNanoToUnixMicro(now().UnixNano())
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[uuid] = &inMemoryItem{
		textures:  textures,
		timestamp: timestamp,
	}
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
	return unixNanoToUnixMicro(now().Add(s.Duration * time.Duration(-1)).UnixNano())
}

func unixNanoToUnixMicro(unixNano int64) int64 {
	return unixNano / 10e5
}
