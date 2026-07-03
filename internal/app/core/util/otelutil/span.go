package otelutil

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"

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

	msg := cfg.msg
	if msg == "" {
		msg = err.Error()
	}

	if cfg.logCTX != nil {
		logMsgs := []any{slog.Any("error", err)}
		for _, lm := range cfg.logMsgs {
			logMsgs = append(logMsgs, lm)
		}

		monitoring.Logger().ErrorContext(cfg.logCTX, msg, logMsgs...)
	}

	span.SetStatus(codes.Error, msg)
	span.RecordError(err)
}
