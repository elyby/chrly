package queue

import (
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"

	"github.com/tevino/abool"
)

var inMemoryStorageGCPeriod = 10 * time.Second
var inMemoryStoragePersistPeriod = time.Minute + 10*time.Second
var now = time.Now

type inMemoryItem struct {
	textures  *mojang.SignedTexturesResponse
	timestamp int64
}

type inMemoryTexturesStorage struct {
	lock    sync.Mutex
	data    map[string]*inMemoryItem
	working *abool.AtomicBool
}

func CreateInMemoryTexturesStorage() *inMemoryTexturesStorage {
	storage := &inMemoryTexturesStorage{
		data: make(map[string]*inMemoryItem),
	}
	storage.Start()

	return storage
}

func (s *inMemoryTexturesStorage) Start() {
	if s.working == nil {
		s.working = abool.New()
	}

	if !s.working.IsSet() {
		go func() {
			time.Sleep(inMemoryStorageGCPeriod)
			// TODO: this can be reimplemented in future with channels, but right now I have no idea how to make it right
			for s.working.IsSet() {
				start := time.Now()
				s.gc()
				time.Sleep(inMemoryStorageGCPeriod - time.Since(start))
			}
		}()
	}

	s.working.Set()
}

func (s *inMemoryTexturesStorage) Stop() {
	s.working.UnSet()
}

func (s *inMemoryTexturesStorage) GetTextures(uuid string) (*mojang.SignedTexturesResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	item, exists := s.data[uuid]
	validRange := getMinimalNotExpiredTimestamp()
	if !exists || validRange > item.timestamp {
		return nil, &ValueNotFound{}
	}

	return item.textures, nil
}

func (s *inMemoryTexturesStorage) StoreTextures(uuid string, textures *mojang.SignedTexturesResponse) {
	var timestamp int64
	if textures != nil {
		decoded := textures.DecodeTextures()
		if decoded == nil {
			panic("unable to decode textures")
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

func (s *inMemoryTexturesStorage) gc() {
	s.lock.Lock()
	defer s.lock.Unlock()

	maxTime := getMinimalNotExpiredTimestamp()
	for uuid, value := range s.data {
		if maxTime > value.timestamp {
			delete(s.data, uuid)
		}
	}
}

func getMinimalNotExpiredTimestamp() int64 {
	return unixNanoToUnixMicro(now().Add(inMemoryStoragePersistPeriod * time.Duration(-1)).UnixNano())
}

func unixNanoToUnixMicro(unixNano int64) int64 {
	return unixNano / 10e5
}
