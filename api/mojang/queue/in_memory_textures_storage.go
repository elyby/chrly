package queue

import (
	"sync"
	"time"

	"github.com/elyby/chrly/api/mojang"

	"github.com/tevino/abool"
)

var inMemoryStorageGCPeriod = time.Second
var inMemoryStoragePersistPeriod = time.Second * 60
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
	return &inMemoryTexturesStorage{
		data: make(map[string]*inMemoryItem),
	}
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
	if !exists || now().Add(inMemoryStoragePersistPeriod*time.Duration(-1)).UnixNano()/10e5 > item.timestamp {
		return nil, &ValueNotFound{}
	}

	return item.textures, nil
}

func (s *inMemoryTexturesStorage) StoreTextures(textures *mojang.SignedTexturesResponse) {
	s.lock.Lock()
	defer s.lock.Unlock()

	decoded := textures.DecodeTextures()
	if decoded == nil {
		panic("unable to decode textures")
	}

	s.data[textures.Id] = &inMemoryItem{
		textures:  textures,
		timestamp: decoded.Timestamp,
	}
}

func (s *inMemoryTexturesStorage) gc() {
	s.lock.Lock()
	defer s.lock.Unlock()

	maxTime := now().Add(inMemoryStoragePersistPeriod*time.Duration(-1)).UnixNano() / 10e5
	for uuid, value := range s.data {
		if maxTime > value.timestamp {
			delete(s.data, uuid)
		}
	}
}
