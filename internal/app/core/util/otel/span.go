package otel

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// @todo add ctx so we can log error
func SpanLogError(span trace.Span, err error, msg string) {
	span.SetStatus(codes.Error, msg)
	span.RecordError(err)
}
