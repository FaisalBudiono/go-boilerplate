package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel/spanattr"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type inputGet interface {
	Context() context.Context
	Actor() *domain.User
	ProductID() string
}

func (srv *Product) Get(req inputGet) (domain.Product, error) {
	ctx, span := srv.tracer.Start(req.Context(), "service: get product")
	defer span.End()

	productID := req.ProductID()
	span.SetAttributes(attribute.String("input.product.id", productID))

	p, err := srv.forceFindProductByID(ctx, domid.ProductID(productID))
	if err != nil {
		return domain.Product{}, err
	}

	if p.PublishedAt != nil {
		return p, nil
	}

	actor := req.Actor()

	if actor == nil {
		return domain.Product{}, tracerr.Wrap(ErrNotFound)
	}
	span.SetAttributes(spanattr.Actor("input.", *actor)...)

	if !slices.Contains(actor.Roles, domain.RoleAdmin) {
		return domain.Product{}, tracerr.Wrap(ErrNotFound)
	}

	return p, nil
}

func (srv *Product) forceFindProductByID(ctx context.Context, id domid.ProductID) (domain.Product, error) {
	p, err := srv.productRepo.FindByID(ctx, srv.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, tracerr.CustomError(ErrNotFound, tracerr.StackTrace(err))
		}

		return domain.Product{}, err
	}

	return p, nil
}
