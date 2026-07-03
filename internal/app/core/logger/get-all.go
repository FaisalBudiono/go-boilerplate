package logger

import (
	"context"
	"log/slog"
	"time"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port/options/activitylog/getall"
)

type reqGetAll interface {
	Context() context.Context
	Actor() domain.Userinfo

	Page() int64
	PerPage() int64

	UserIDFlags() []string
	ActionFlags() []actlog.Action

	CreatedFrom() time.Time
	CreatedTo() time.Time
}

func (srv *Logger) GetAll(
	req reqGetAll,
) ([]domain.ActivityLogEagerLoad, domain.Pagination, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.sName("get-all"))
	defer span.End()

	actor := req.Actor()
	page := req.Page()
	perPage := req.PerPage()
	userIDFlags := req.UserIDFlags()
	actionFlags := req.ActionFlags()
	createdFrom := req.CreatedFrom()
	createdTo := req.CreatedTo()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.Any("actor", actor),
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Any("userIDFlags", userIDFlags),
		slog.Any("actionFlags", actionFlags),
		slog.String("createdFrom", createdFrom.Format(time.RFC3339)),
		slog.String("createdTo", createdTo.Format(time.RFC3339)),
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

	logs, total, err := srv.activityLogRepo.GetPaginated(
		ctx, srv.db, page, perPage,
		getall.WithUserIDFlags(userIDFlags...),
		getall.WithActionFlags(actionFlags...),
		getall.WithCreatedFrom(createdFrom),
		getall.WithCreatedTo(createdTo),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get activity logs"),
		)
		return nil, emptyPage, err
	}

	eagerLogs, err := srv.eagerLoader.All(ctx, logs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load activity logs"),
		)
		return nil, emptyPage, err
	}

	return eagerLogs, domain.NewPagination(page, perPage, total), nil
}
