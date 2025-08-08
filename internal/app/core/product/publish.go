package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"context"
	"log/slog"
	"slices"
)

type inputPublish interface {
	Context() context.Context
	Actor() domain.User
	IsPublish() bool
	ProductID() string
}

func (srv *Product) Publish(req inputPublish) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.Product.Publish")
	defer span.End()

	actor := req.Actor()
	isPublish := req.IsPublish()
	productID := req.ProductID()

	logVals := []any{slog.Bool("isPublish", isPublish), slog.String("product.id", productID)}
	logVals = append(logVals, logutil.SlogActor(actor)...)
	monitoring.Logger().InfoContext(ctx, "input", logVals...)

	if !slices.Contains(actor.Roles, domain.RoleAdmin) {
		return domain.Product{}, ErrNotEnoughPermission
	}

	p, err := srv.forceFindProductByID(ctx, domid.ProductID(productID))
	if err != nil {
		return domain.Product{}, err
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, err
	}
	defer tx.Rollback()

	p, err = srv.productRepo.Publish(ctx, tx, p, isPublish)
	if err != nil {
		return domain.Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Product{}, err
	}

	return p, nil
}
