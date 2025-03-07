package product

import (
	"FaisalBudiono/go-boilerplate/internal/domain"
	"FaisalBudiono/go-boilerplate/internal/otel/spanattr"
	"context"
	"slices"
	"strings"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/attribute"
)

type inputSave interface {
	Context() context.Context
	Actor() domain.User
	Name() string
	Price() int64
}

func (srv *Product) Save(req inputSave) (domain.Product, error) {
	ctx, span := srv.tracer.Start(req.Context(), "service: save product")
	defer span.End()

	actor := req.Actor()
	name := strings.Trim(req.Name(), " ")
	price := req.Price()

	span.SetAttributes(
		attribute.String("input.name", name),
		attribute.Int64("input.price", price),
	)
	span.SetAttributes(spanattr.Actor("input.", actor)...)

	if !slices.Contains(actor.Roles, domain.RoleAdmin) {
		return domain.Product{}, tracerr.Wrap(ErrNotEnoughPermission)
	}

	if name == "" {
		return domain.Product{}, tracerr.Wrap(ErrEmptyName)
	}

	if price < 0 {
		return domain.Product{}, tracerr.Wrap(ErrNegativePrice)
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, tracerr.Wrap(err)
	}
	defer tx.Rollback()

	p, err := srv.productSaver.Save(ctx, tx, name, price)
	if err != nil {
		return domain.Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Product{}, tracerr.Wrap(err)
	}

	return p, nil
}
