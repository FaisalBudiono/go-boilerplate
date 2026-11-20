package mon

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	gTracer      trace.Tracer
	gLogger      *slog.Logger
	stdoutLogger *slog.Logger
)

func init() {
	gTracer = noop.Tracer{}
	gLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	stdoutLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func SetUp(tracer trace.Tracer, logger *slog.Logger) {
	gTracer = tracer
	gLogger = logger

	stdoutOnly := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLogLevel(),
	})

	stdoutLogger = slog.New(newTeeHandler(stdoutOnly, logger.Handler()))
}

func Logger() *slog.Logger { return gLogger }
func Tracer() trace.Tracer { return gTracer }
func Stdout() *slog.Logger { return stdoutLogger }

func defaultLogLevel() slog.Leveler {
	switch app.ENV().Log.Level {
	case app.LogLevelDebug:
		return slog.LevelDebug

	case app.LogLevelInfo:
		return slog.LevelInfo

	case app.LogLevelWarn:
		return slog.LevelWarn

	case app.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// teeHandler duplicates log records to multiple handlers.
type teeHandler struct {
	handlers []slog.Handler
}

func newTeeHandler(handlers ...slog.Handler) *teeHandler {
	return &teeHandler{handlers: handlers}
}

func (t *teeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Return true if AT LEAST ONE handler is enabled for this level
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (t *teeHandler) Handle(ctx context.Context, rec slog.Record) error {
	var errs []error
	for _, h := range t.handlers {
		if h.Enabled(ctx, rec.Level) {
			// Use rec.Clone() to prevent handlers from accidentally sharing/mutating state
			if err := h.Handle(ctx, rec.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range t.handlers {
		newHandlers = append(newHandlers, h.WithAttrs(attrs))
	}
	return newTeeHandler(newHandlers...)
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	var newHandlers []slog.Handler
	for _, h := range t.handlers {
		newHandlers = append(newHandlers, h.WithGroup(name))
	}
	return newTeeHandler(newHandlers...)
}
