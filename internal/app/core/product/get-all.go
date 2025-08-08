package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout/options/product/optproduct"
	"context"
	"log/slog"
	"slices"
)

type inputGetAll interface {
	Context() context.Context
	Actor() *domain.User

	Page() int64
	PerPage() int64

	CMSAcces() bool
}

func (srv *Product) GetAll(req inputGetAll) ([]domain.Product, domain.Pagination, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.Product.GetAll")
	defer span.End()

	actor := req.Actor()
	page := req.Page()
	perPage := req.PerPage()
	cmsAccess := req.CMSAcces()

	logVals := []any{
		slog.Int64("page", page),
		slog.Int64("perPage", perPage),
		slog.Bool("cmsAccess", cmsAccess),
	}
	if actor != nil {
		logVals = append(logVals, logutil.SlogActor(*actor)...)
	}
	monitoring.Logger().InfoContext(ctx, "input", logVals...)

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	offset := (page - 1) * perPage

	isNotAdmin := actor == nil || !slices.Contains(actor.Roles, domain.RoleAdmin)
	if isNotAdmin {
		products, total, err := srv.productRepo.GetAll(ctx, srv.db, offset, perPage)
		if err != nil {
			return nil, domain.Pagination{}, err
		}

		return products, domain.NewPagination(page, perPage, total), nil
	}

	span.AddEvent("fetched product for admin")

	products, total, err := srv.productRepo.GetAll(
		ctx,
		srv.db,
		offset,
		perPage,
		optproduct.WithShowFlag(cmsAccess),
	)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	return products, domain.NewPagination(page, perPage, total), nil
}
