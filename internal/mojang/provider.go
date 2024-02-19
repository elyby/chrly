package mojang

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/brunomvsouza/singleflight"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"
)

const ScopeName = "ely.by/chrly/internal/mojang"

var InvalidUsername = errors.New("the username passed doesn't meet Mojang's requirements")

// https://help.minecraft.net/hc/en-us/articles/4408950195341#h_01GE5JX1Z0CZ833A7S54Y195KV
var allowedUsernamesRegex = regexp.MustCompile(`(?i)^[0-9a-z_]{3,16}$`)

type UuidsProvider interface {
	GetUuid(ctx context.Context, username string) (*ProfileInfo, error)
}

type TexturesProvider interface {
	GetTextures(ctx context.Context, uuid string) (*ProfileResponse, error)
}

func NewMojangTexturesProvider(
	uuidsProvider UuidsProvider,
	texturesProvider TexturesProvider,
) (*MojangTexturesProvider, error) {
	meter, err := newProviderMetrics(otel.GetMeterProvider().Meter(ScopeName))
	if err != nil {
		return nil, err
	}

	return &MojangTexturesProvider{
		UuidsProvider:    uuidsProvider,
		TexturesProvider: texturesProvider,
		metrics:          meter,
	}, nil
}

type MojangTexturesProvider struct {
	UuidsProvider
	TexturesProvider

	metrics *providerMetrics
	group   singleflight.Group[string, *ProfileResponse]
}

func (p *MojangTexturesProvider) GetForUsername(ctx context.Context, username string) (*ProfileResponse, error) {
	if !allowedUsernamesRegex.MatchString(username) {
		return nil, InvalidUsername
	}

	username = strings.ToLower(username)

	result, err, shared := p.group.Do(username, func() (*ProfileResponse, error) {
		profile, err := p.UuidsProvider.GetUuid(ctx, username)
		if err != nil {
			return nil, err
		}

		if profile == nil {
			return nil, nil
		}

		return p.TexturesProvider.GetTextures(ctx, profile.Id)
	})

	p.recordMetrics(ctx, shared, result, err)

	return result, err
}

func (p *MojangTexturesProvider) recordMetrics(ctx context.Context, shared bool, result *ProfileResponse, err error) {
	if shared {
		p.metrics.Shared.Add(ctx, 1)
	}

	if err != nil {
		p.metrics.Failed.Add(ctx, 1)
		return
	}

	if result != nil {
		p.metrics.Found.Add(ctx, 1)
	} else {
		p.metrics.Missed.Add(ctx, 1)
	}
}

type NilProvider struct {
}

func (*NilProvider) GetForUsername(ctx context.Context, username string) (*ProfileResponse, error) {
	return nil, nil
}

func newProviderMetrics(meter metric.Meter) (*providerMetrics, error) {
	m := &providerMetrics{}
	var errors, err error

	m.Found, err = meter.Int64Counter(
		"results.found",
		metric.WithDescription(""), // TODO: description
	)
	errors = multierr.Append(errors, err)

	m.Missed, err = meter.Int64Counter(
		"results.missed",
		metric.WithDescription(""), // TODO: description
	)
	errors = multierr.Append(errors, err)

	m.Failed, err = meter.Int64Counter(
		"results.failed",
		metric.WithDescription(""), // TODO: description
	)
	errors = multierr.Append(errors, err)

	m.Shared, err = meter.Int64Counter(
		"singleflight.shared",
		metric.WithDescription(""), // TODO: description
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type providerMetrics struct {
	Found  metric.Int64Counter
	Missed metric.Int64Counter
	Failed metric.Int64Counter
	Shared metric.Int64Counter
}
