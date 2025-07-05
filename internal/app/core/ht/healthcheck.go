package ht

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"context"
	"database/sql"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/codes"
)

type Healthcheck struct {
	db *sql.DB
}

type inputHealthcheck interface {
	Context() context.Context
}

func (srv *Healthcheck) Healthcheck(req inputHealthcheck) error {
	ctx, span := monitorings.Tracer().Start(req.Context(), "core.ht.healthcheck")
	defer span.End()

	err := srv.db.PingContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "sql error")
		span.RecordError(err)

		return tracerr.Wrap(err)
	}

	return nil
}

func New(db *sql.DB) *Healthcheck {
	return &Healthcheck{
		db: db,
	}
}
