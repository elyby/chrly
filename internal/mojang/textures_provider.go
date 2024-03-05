package mojang

import (
	"context"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
)

type MojangApiTexturesProviderFunc func(ctx context.Context, uuid string, signed bool) (*ProfileResponse, error)

func NewMojangApiTexturesProvider(endpoint MojangApiTexturesProviderFunc) (*MojangApiTexturesProvider, error) {
	metrics, err := newMojangApiTexturesProviderMetrics(otel.GetMeterProvider().Meter(ScopeName))
	if err != nil {
		return nil, err
	}

	return &MojangApiTexturesProvider{
		MojangApiTexturesEndpoint: endpoint,
		metrics:                   metrics,
	}, nil
}

type MojangApiTexturesProvider struct {
	MojangApiTexturesEndpoint MojangApiTexturesProviderFunc
	metrics                   *mojangApiTexturesProviderMetrics
}

func (p *MojangApiTexturesProvider) GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error) {
	p.metrics.Requests.Add(ctx, 1)

	return p.MojangApiTexturesEndpoint(ctx, uuid, true)
}

// Perfectly there should be an object with provider and cache implementation,
// but I decided not to introduce a layer and just implement cache in place.
type TexturesProviderWithInMemoryCache struct {
	provider TexturesProvider
	once     sync.Once
	cache    *ttlcache.Cache[string, *ProfileResponse]
	metrics  *texturesProviderWithInMemoryCacheMetrics
}

func NewTexturesProviderWithInMemoryCache(provider TexturesProvider) (*TexturesProviderWithInMemoryCache, error) {
	metrics, err := newTexturesProviderWithInMemoryCacheMetrics(otel.GetMeterProvider().Meter(ScopeName))
	if err != nil {
		return nil, err
	}

	return &TexturesProviderWithInMemoryCache{
		provider: provider,
		cache: ttlcache.New[string, *ProfileResponse](
			ttlcache.WithDisableTouchOnHit[string, *ProfileResponse](),
			// I'm aware of ttlcache.WithLoader(), but it doesn't allow to return an error
		),
		metrics: metrics,
	}, nil
}

func (s *TexturesProviderWithInMemoryCache) GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error) {
	item := s.cache.Get(uuid)
	// Don't check item.IsExpired() since Get function is already did this check
	if item != nil {
		s.metrics.Hits.Add(ctx, 1)
		return item.Value(), nil
	}

	s.metrics.Misses.Add(ctx, 1)

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

func newMojangApiTexturesProviderMetrics(meter metric.Meter) (*mojangApiTexturesProviderMetrics, error) {
	m := &mojangApiTexturesProviderMetrics{}
	var errors, err error

	m.Requests, err = meter.Int64Counter(
		"textures.request.sent",
		metric.WithDescription("Number of textures requests sent to Mojang API"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type mojangApiTexturesProviderMetrics struct {
	Requests metric.Int64Counter
}

func newTexturesProviderWithInMemoryCacheMetrics(meter metric.Meter) (*texturesProviderWithInMemoryCacheMetrics, error) {
	m := &texturesProviderWithInMemoryCacheMetrics{}
	var errors, err error

	m.Hits, err = meter.Int64Counter(
		"textures.cache.hit",
		metric.WithDescription("Number of Mojang textures found in the local cache"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.Misses, err = meter.Int64Counter(
		"textures.cache.miss",
		metric.WithDescription("Number of Mojang textures missing from local cache"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type texturesProviderWithInMemoryCacheMetrics struct {
	Hits   metric.Int64Counter
	Misses metric.Int64Counter
}
