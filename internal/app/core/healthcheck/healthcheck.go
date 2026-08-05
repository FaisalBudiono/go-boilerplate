package healthcheck

import (
	"context"
	"fmt"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

func New(
	db port.DBTX,
) *Healthcheck {
	return &Healthcheck{
		db: db,
	}
}

type Healthcheck struct {
	db port.DBTX
}

type reqCheck interface {
	Context() context.Context
}

func (srv *Healthcheck) Check(req reqCheck) []domain.HealthcheckDep {
	ctx, span := mon.Tracer().Start(req.Context(), "core.healthcheck.check")
	defer span.End()

	return []domain.HealthcheckDep{
		srv.checkDB(ctx),
	}
}

func (srv *Healthcheck) checkDB(ctx context.Context) domain.HealthcheckDep {
	ctx, span := mon.Tracer().Start(ctx, "core.healthcheck.db")
	defer span.End()

	health := domain.NewHealthcheckDep("database", false, "")

	query := "SELECT 1"

	var result int
	err := srv.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("database health check failed"),
		)

		health.Status = fmt.Sprintf("unhealthy: %s", err.Error())
		health.Ok = false

		return health
	}

	health.Status = "healthy"
	health.Ok = true

	return health
}
