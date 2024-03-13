package mojang

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/brunomvsouza/singleflight"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/multierr"

	"ely.by/chrly/internal/otel"
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
	meter, err := newProviderMetrics(otel.GetMeter())
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
		var profile *ProfileInfo
		var textures *ProfileResponse
		var err error

		defer p.recordMetrics(ctx, profile, textures, err)

		profile, err = p.UuidsProvider.GetUuid(ctx, username)
		if err != nil {
			return nil, err
		}

		if profile == nil {
			return nil, nil
		}

		textures, err = p.TexturesProvider.GetTextures(ctx, profile.Id)

		return textures, err
	})

	if shared {
		p.metrics.Shared.Add(ctx, 1)
	}

	return result, err
}

func (p *MojangTexturesProvider) recordMetrics(ctx context.Context, profile *ProfileInfo, textures *ProfileResponse, err error) {
	if err != nil {
		p.metrics.Failed.Add(ctx, 1)
		return
	}

	if profile == nil {
		p.metrics.UsernameMissed.Add(ctx, 1)
		p.metrics.TextureMissed.Add(ctx, 1)

		return
	}

	p.metrics.UsernameFound.Add(ctx, 1)
	if textures != nil {
		p.metrics.TextureFound.Add(ctx, 1)
	} else {
		p.metrics.TextureMissed.Add(ctx, 1)
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

	m.UsernameFound, err = meter.Int64Counter(
		"mojang.provider.username_found",
		metric.WithDescription("Number of queries for which username was found"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.UsernameMissed, err = meter.Int64Counter(
		"chrly.mojang.provider.username_missed",
		metric.WithDescription("Number of queries for which username was not found"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.TextureFound, err = meter.Int64Counter(
		"chrly.mojang.provider.textures_found",
		metric.WithDescription("Number of queries for which textures were successfully found"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.TextureMissed, err = meter.Int64Counter(
		"chrly.mojang.provider.textures_missed",
		metric.WithDescription("Number of queries for which no textures were found"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.Failed, err = meter.Int64Counter(
		"chrly.mojang.provider.failed",
		metric.WithDescription("Number of requests that ended in an error"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	m.Shared, err = meter.Int64Counter(
		"chrly.mojang.provider.singleflight.shared",
		metric.WithDescription("Number of requests that are already being processed in another thread"),
		metric.WithUnit("1"),
	)
	errors = multierr.Append(errors, err)

	return m, errors
}

type providerMetrics struct {
	UsernameFound  metric.Int64Counter
	UsernameMissed metric.Int64Counter
	TextureFound   metric.Int64Counter
	TextureMissed  metric.Int64Counter
	Failed         metric.Int64Counter
	Shared         metric.Int64Counter
}
