package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"slices"

	"go.opentelemetry.io/otel/attribute"
)

type inputGetAll interface {
	Context() context.Context
	Actor() *domain.User

	Page() int64
	PerPage() int64

	CMSAcces() bool
}

func (srv *Product) GetAll(req inputGetAll) ([]domain.Product, domain.Pagination, error) {
	ctx, span := monitorings.Tracer().Start(req.Context(), "service: get all product")
	defer span.End()

	actor := req.Actor()
	page := req.Page()
	perPage := req.PerPage()
	showAll := req.CMSAcces()

	span.SetAttributes(attribute.Int64("input.page", page))
	span.SetAttributes(attribute.Int64("input.perPage", perPage))
	span.SetAttributes(attribute.Bool("input.cmsAccess", showAll))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	offset := (page - 1) * perPage

	isNotAdmin := actor == nil || !slices.Contains(actor.Roles, domain.RoleAdmin)
	if isNotAdmin {
		products, total, err := srv.productRepo.GetAll(ctx, srv.db, false, offset, perPage)
		if err != nil {
			return nil, domain.Pagination{}, err
		}

		return products, domain.NewPagination(page, perPage, total), nil
	}

	span.AddEvent("fetched product for admin")

	products, total, err := srv.productRepo.GetAll(ctx, srv.db, showAll, offset, perPage)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	return products, domain.NewPagination(page, perPage, total), nil
}
