package monitorings

import (
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	gTracer trace.Tracer
	gLogger *slog.Logger
)

func init() {
	gTracer = noop.Tracer{}
	gLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func SetUp(tracer trace.Tracer, logger *slog.Logger) {
	gTracer = tracer
	gLogger = logger
}

func Logger() *slog.Logger {
	return gLogger
}

func Tracer() trace.Tracer {
	return gTracer
}
