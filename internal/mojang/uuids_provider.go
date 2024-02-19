package mojang

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
)

type MojangUuidsStorage interface {
	// The second argument must be returned as a incoming username in case,
	// when cached result indicates that there is no Mojang user with provided username
	GetUuidForMojangUsername(ctx context.Context, username string) (foundUuid string, foundUsername string, err error)
	// An empty uuid value can be passed if the corresponding account has not been found
	StoreMojangUuid(ctx context.Context, username string, uuid string) error
}

func NewUuidsProviderWithCache(o UuidsProvider, s MojangUuidsStorage) (*UuidsProviderWithCache, error) {
	metrics, err := newUuidsProviderWithCacheMetrics(otel.GetMeterProvider().Meter(ScopeName))
	if err != nil {
		return nil, err
	}

	return &UuidsProviderWithCache{
		Provider: o,
		Storage:  s,
		metrics:  metrics,
	}, nil
}

type UuidsProviderWithCache struct {
	Provider UuidsProvider
	Storage  MojangUuidsStorage

	metrics *uuidsProviderWithCacheMetrics
}

func (p *UuidsProviderWithCache) GetUuid(ctx context.Context, username string) (*ProfileInfo, error) {
	var uuid, foundUsername string
	var err error
	defer p.recordMetrics(ctx, uuid, foundUsername, err)

	uuid, foundUsername, err = p.Storage.GetUuidForMojangUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if foundUsername != "" {
		if uuid != "" {
			return &ProfileInfo{Id: uuid, Name: foundUsername}, nil
		}

		return nil, nil
	}

	profile, err := p.Provider.GetUuid(ctx, username)
	if err != nil {
		return nil, err
	}

	freshUuid := ""
	wellCasedUsername := username
	if profile != nil {
		freshUuid = profile.Id
		wellCasedUsername = profile.Name
	}

	_ = p.Storage.StoreMojangUuid(ctx, wellCasedUsername, freshUuid)

	return profile, nil
}

func (p *UuidsProviderWithCache) recordMetrics(ctx context.Context, uuid string, username string, err error) {
	if err != nil {
		return
	}

	if username != "" {
		p.metrics.Hits.Add(ctx, 1)
	} else {
		p.metrics.Misses.Add(ctx, 1)
	}
}

func newUuidsProviderWithCacheMetrics(meter metric.Meter) (*uuidsProviderWithCacheMetrics, error) {
	m := &uuidsProviderWithCacheMetrics{}
	var errors, err error

	m.Hits, err = meter.Int64Counter(
		"uuids.cache.hit",
		metric.WithDescription(""), // TODO: write description
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.Misses, err = meter.Int64Counter(
		"uuids.cache.miss",
		metric.WithDescription(""), // TODO: write description
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type uuidsProviderWithCacheMetrics struct {
	Hits   metric.Int64Counter
	Misses metric.Int64Counter
}
