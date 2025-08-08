package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"log/slog"
	"slices"
	"strings"
)

type inputSave interface {
	Context() context.Context
	Actor() domain.User
	Name() string
	Price() int64
}

func (srv *Product) Save(req inputSave) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.Product.Save")
	defer span.End()

	actor := req.Actor()
	name := strings.Trim(req.Name(), " ")
	price := req.Price()

	logVals := []any{slog.String("name", name), slog.Int64("price", price)}
	logVals = append(logVals, logutil.SlogActor(actor)...)
	monitoring.Logger().InfoContext(ctx, "input", logVals...)

	if !slices.Contains(actor.Roles, domain.RoleAdmin) {
		return domain.Product{}, ErrNotEnoughPermission
	}

	if name == "" {
		return domain.Product{}, ErrEmptyName
	}

	if price < 0 {
		return domain.Product{}, ErrNegativePrice
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, err
	}
	defer tx.Rollback()

	p, err := srv.productRepo.Save(ctx, tx, name, price)
	if err != nil {
		return domain.Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Product{}, err
	}

	return p, nil
}
