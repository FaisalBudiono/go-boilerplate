package otel

import (
	"fmt"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func SpanLogError(span trace.Span, err error, msg string) {
	i := 0
	stacks := tracerr.StackTrace(err)

	for len(stacks) > 0 {
		for _, s := range stacks {
			span.SetAttributes(
				attribute.String(fmt.Sprintf("errorTrace%d", i), s.String()),
			)

			i++
		}

		err = tracerr.Unwrap((err))
		stacks = tracerr.StackTrace(err)
	}

	span.SetStatus(codes.Error, msg)
	span.RecordError(err)
}
