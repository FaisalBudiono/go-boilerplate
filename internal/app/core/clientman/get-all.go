package clientman

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
	getallopt "komdigi-immigration/internal/app/port/options/clientcred/getall"
)

const (
	defaultPage    int64 = 1
	defaultPerPage int64 = 10
)

type reqGetAll interface {
	Context() context.Context
	Actor() domain.Userinfo

	Page() *int64
	PerPage() *int64
	UserIDFlags() []string
	SearchClientID() string
}

func (srv *ClientManager) GetAll(
	req reqGetAll,
) ([]domain.ClientCredentialEagerLoad, domain.Pagination, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("get-all"))
	defer span.End()

	pPage := req.Page()
	pPerPage := req.PerPage()
	actor := req.Actor()
	userIDFlags := req.UserIDFlags()
	searchClientID := req.SearchClientID()

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
		slog.Any("userIDFlags", userIDFlags),
		slog.String("searchClientID", searchClientID),
		slog.Any("actor", actor),
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

	uniqueUserIDs := sliceutil.Unique(userIDFlags)

	ccs, total, err := srv.clientCredRepo.GetPaginated(
		ctx, srv.db, page, perPage,
		getallopt.WithUserIDs(uniqueUserIDs...),
		getallopt.WithSearchClientID(searchClientID),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get client credentials"),
		)
		return nil, emptyPage, err
	}

	eagerRes, err := srv.eagerLoad.All(ctx, ccs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load client credentials"),
		)
		return nil, emptyPage, err
	}

	return eagerRes, domain.NewPagination(page, perPage, total), nil
}
