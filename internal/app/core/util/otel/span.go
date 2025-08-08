package otel

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type spanLogConfig struct {
	logCTX context.Context

	logMsgs []slog.Attr
	msg     string
}

func newSpanLogConfig() *spanLogConfig {
	return &spanLogConfig{
		logMsgs: make([]slog.Attr, 0),
	}
}

type spanLogOption func(*spanLogConfig)

func WithErrorLog(ctx context.Context, msgs ...slog.Attr) spanLogOption {
	return func(opt *spanLogConfig) {
		opt.logCTX = ctx
		opt.logMsgs = msgs
	}
}

func WithMessage(msg string) spanLogOption {
	return func(opt *spanLogConfig) {
		opt.msg = msg
	}
}

func SpanLogError(span trace.Span, err error, opts ...spanLogOption) {
	cfg := newSpanLogConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.logCTX != nil {
		logMsgs := []any{slog.Any("error", err)}
		for _, lm := range logMsgs {
			logMsgs = append(logMsgs, lm)
		}

		monitoring.Logger().ErrorContext(cfg.logCTX, "error", logMsgs...)
	}

	msg := cfg.msg
	if msg == "" {
		msg = err.Error()
	}

	span.SetStatus(codes.Error, msg)
	span.RecordError(err)
}
