package imig

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
)

type reqIMEIHistory interface {
	Context() context.Context
	Actor() domain.Userinfo

	IMEI() string
	Page() int64
	PerPage() int64
}

const (
	defaultPage    int64 = 1
	defaultPerPage int64 = 10
)

func (srv *Imig) IMEIHistory(
	req reqIMEIHistory,
) ([]domain.ImmigrationClearanceLogEagerLoad, domain.Pagination, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.sName("imei-history"))
	defer span.End()

	actor := req.Actor()
	imei := req.IMEI()
	page := req.Page()
	perPage := req.PerPage()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.Any("actor", actor),
		slog.String("imei", imei),
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
	)

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}

	emptyPage := domain.Pagination{}
	if !actor.User.HasRoles(domain.RoleAdmin) {
		return nil, emptyPage, ErrPermissionDenied
	}

	icLogs, total, err := srv.icLogRepo.GetIMEIHistoryPaginated(ctx, srv.db, imei, page, perPage)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get imei history"),
		)
		return nil, emptyPage, err
	}

	eagerLogs, err := srv.eagerLoader.All(ctx, icLogs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load"),
		)
		return nil, emptyPage, err
	}

	return eagerLogs, domain.NewPagination(page, perPage, total), nil
}
