package otel

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func NewLogger(appName string) *slog.Logger {
	return otelslog.NewLogger(appName)
}
