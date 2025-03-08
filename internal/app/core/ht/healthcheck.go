package ht

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Healthcheck struct {
	db *sql.DB

	tracer trace.Tracer
	logger *slog.Logger
}

type inputHealthcheck interface {
	Context() context.Context
}

func (srv *Healthcheck) Healthcheck(req inputHealthcheck) error {
	ctx, span := srv.tracer.Start(req.Context(), "service: healthcheck")
	defer span.End()

	err := srv.db.PingContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "sql error")
		span.RecordError(err)

		return tracerr.Wrap(err)
	}

	return nil
}

func New(
	db *sql.DB,
	tracer trace.Tracer,
	logger *slog.Logger,
) *Healthcheck {
	return &Healthcheck{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}
