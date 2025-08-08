package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
)

type inputGet interface {
	Context() context.Context
	Actor() *domain.User
	ProductID() string
}

func (srv *Product) Get(req inputGet) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.Product.Get")
	defer span.End()

	productID := req.ProductID()
	actor := req.Actor()

	logVals := []any{slog.String("product.id", productID)}
	if actor != nil {
		logVals = append(logVals, logutil.SlogActor(*actor)...)
	}
	monitoring.Logger().InfoContext(ctx, "input", logVals...)

	p, err := srv.forceFindProductByID(ctx, domid.ProductID(productID))
	if err != nil {
		return domain.Product{}, err
	}

	if p.PublishedAt != nil {
		return p, nil
	}

	isNotAdmin := actor == nil || !slices.Contains(actor.Roles, domain.RoleAdmin)
	if isNotAdmin {
		return domain.Product{}, ErrNotFound
	}

	return p, nil
}

func (srv *Product) forceFindProductByID(ctx context.Context, id domid.ProductID) (domain.Product, error) {
	p, err := srv.productRepo.FindByID(ctx, srv.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, errors.Join(ErrNotFound, err)
		}

		return domain.Product{}, err
	}

	return p, nil
}
