package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/util/perm"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"FaisalBudiono/go-boilerplate/internal/otel/spanattr"
	"context"
	"database/sql"
	"errors"

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

	pID := req.ProductID()
	span.SetAttributes(attribute.String("input.product.id", pID))

	p, err := srv.forceFindProductByID(ctx, req.ProductID())
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

	if !perm.IsAdmin(*actor) {
		return domain.Product{}, tracerr.Wrap(ErrNotFound)
	}

	return p, nil
}

func (srv *Product) forceFindProductByID(ctx context.Context, productID string) (domain.Product, error) {
	p, err := srv.productIDFinder.FindByID(ctx, srv.db, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, tracerr.CustomError(ErrNotFound, tracerr.StackTrace(err))
		}

		return domain.Product{}, err
	}

	return p, nil
}
