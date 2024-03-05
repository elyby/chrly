package otel

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/agoda-com/opentelemetry-go/otelslog"
	logsOtel "github.com/agoda-com/opentelemetry-logs-go"
	logsAutoconfig "github.com/agoda-com/opentelemetry-logs-go/autoconfigure/sdk/logs"
	"github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.4.0"

	"ely.by/chrly/internal/version"
)

func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		shutdownFuncs = nil

		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up resource
	res, err := newResource(ctx)
	if err != nil {
		handleErr(err)
		return
	}

	// Set up logs provider
	logsProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, logsProvider.Shutdown)
	logsOtel.SetLoggerProvider(logsProvider)

	otelSlog := slog.New(otelslog.NewOtelHandler(logsProvider, &otelslog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(otelSlog)

	// Set up trace provider
	tracerProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider
	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	err = runtimeMetrics.Start(runtimeMetrics.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		handleErr(err)
		return
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("chrly"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*logs.LoggerProvider, error) {
	return logsAutoconfig.NewLoggerProvider(ctx, logsAutoconfig.WithResource(res)), nil
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	), nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, err
	}

	return metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	), nil
}
