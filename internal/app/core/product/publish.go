package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel/spanattr"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
	"slices"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type inputPublish interface {
	Context() context.Context
	Actor() domain.User
	IsPublish() bool
	ProductID() string
}

func (srv *Product) Publish(req inputPublish) (domain.Product, error) {
	ctx, span := monitorings.Tracer().Start(req.Context(), "service: publish product")
	defer span.End()

	actor := req.Actor()
	isPublish := req.IsPublish()
	productID := req.ProductID()

	span.SetAttributes(
		attribute.Bool("input.isPublish", isPublish),
		attribute.String("input.product.id", productID),
	)
	span.SetAttributes(spanattr.Actor("input.", actor)...)

	if !slices.Contains(actor.Roles, domain.RoleAdmin) {
		return domain.Product{}, tracerr.Wrap(ErrNotEnoughPermission)
	}

	p, err := srv.forceFindProductByID(ctx, domid.ProductID(productID))
	if err != nil {
		return domain.Product{}, err
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, tracerr.Wrap(err)
	}
	defer tx.Rollback()

	p, err = srv.productRepo.Publish(ctx, tx, p, isPublish)
	if err != nil {
		return domain.Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Product{}, tracerr.Wrap(err)
	}

	return p, nil
}
