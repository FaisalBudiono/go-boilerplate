package user

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
	getallopt "komdigi-immigration/internal/app/port/options/user/getall"
)

type reqGetAll interface {
	Context() context.Context
	Page() *int64
	PerPage() *int64
	RoleFlags() []domain.Role
	Actor() domain.Userinfo
}

func (srv *User) GetAll(
	req reqGetAll,
) ([]domain.UserEagerLoad, domain.Pagination, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.user.get-all")
	defer span.End()

	pPage := req.Page()
	pPerPage := req.PerPage()
	roleFlags := req.RoleFlags()
	actor := req.Actor()

	page := defaultPage
	if pPage != nil {
		page = *pPage
	}

	perPage := defaultPerPage
	if pPerPage != nil {
		perPage = *pPerPage
	}

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Any("roleFlags", roleFlags),
		slog.Any("actor", actor),
	)

	emptyPagination := domain.Pagination{}

	if !actor.User.HasRoles(domain.RoleAdmin) {
		return nil, emptyPagination, ErrPermissionDenied
	}

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}

	cleanRoles := make([]domain.Role, 0)
	for _, role := range roleFlags {
		if role.IsValid() {
			cleanRoles = append(cleanRoles, role)
		}
	}
	cleanRoles = sliceutil.Unique(cleanRoles)

	users, total, err := srv.userRepo.GetPaginated(
		ctx, srv.db, page, perPage,
		getallopt.WithRoleFlags(cleanRoles...),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get users"),
		)
		return nil, emptyPagination, err
	}

	eagerUsers, err := srv.eagerLoad.All(ctx, users)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load users"),
		)
		return nil, emptyPagination, err
	}

	return eagerUsers, domain.NewPagination(page, perPage, total), nil
}
