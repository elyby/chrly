package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const Scope = "ely.by/chrly"

func GetMeter(opts ...metric.MeterOption) metric.Meter {
	return otel.GetMeterProvider().Meter(Scope, opts...)
}

func GetTracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(Scope, opts...)
}
