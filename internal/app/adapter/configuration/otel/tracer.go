package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func NewTracer(appName string) trace.Tracer {
	return otel.Tracer(appName)
}
