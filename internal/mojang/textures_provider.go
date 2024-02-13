package mojang

import (
	"context"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type MojangApiTexturesProvider struct {
	MojangApiTexturesEndpoint func(ctx context.Context, uuid string, signed bool) (*ProfileResponse, error)
}

func (p *MojangApiTexturesProvider) GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error) {
	return p.MojangApiTexturesEndpoint(ctx, uuid, true)
}

// Perfectly there should be an object with provider and cache implementation,
// but I decided not to introduce a layer and just implement cache in place.
type TexturesProviderWithInMemoryCache struct {
	provider TexturesProvider
	once     sync.Once
	cache    *ttlcache.Cache[string, *ProfileResponse]
}

func NewTexturesProviderWithInMemoryCache(provider TexturesProvider) *TexturesProviderWithInMemoryCache {
	storage := &TexturesProviderWithInMemoryCache{
		provider: provider,
		cache: ttlcache.New[string, *ProfileResponse](
			ttlcache.WithDisableTouchOnHit[string, *ProfileResponse](),
			// I'm aware of ttlcache.WithLoader(), but it doesn't allow to return an error
		),
	}

	return storage
}

func (s *TexturesProviderWithInMemoryCache) GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error) {
	item := s.cache.Get(uuid)
	// Don't check item.IsExpired() since Get function is already did this check
	if item != nil {
		return item.Value(), nil
	}

	result, err := s.provider.GetTextures(ctx, uuid)
	if err != nil {
		return nil, err
	}

	s.cache.Set(uuid, result, time.Minute)
	// Call it only after first set so GC will work more often
	s.startGcOnce()

	return result, nil
}

func (s *TexturesProviderWithInMemoryCache) StopGC() {
	// If you call the Stop() on a non-started GC, the process will hang trying to close the uninitialized channel
	s.startGcOnce()
	s.cache.Stop()
}

func (s *TexturesProviderWithInMemoryCache) startGcOnce() {
	s.once.Do(func() {
		go s.cache.Start()
	})
}
