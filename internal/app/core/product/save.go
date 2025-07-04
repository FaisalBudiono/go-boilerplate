package product

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/ztrue/tracerr"
)

type inputSave interface {
	Context() context.Context
	Actor() domain.User
	Name() string
	Price() int64
}

func (srv *Product) Save(req inputSave) (domain.Product, error) {
	ctx, span := monitorings.Tracer().Start(req.Context(), "core.product.save")
	defer span.End()

	actor := req.Actor()
	name := strings.Trim(req.Name(), " ")
	price := req.Price()

	logVals := []any{slog.String("name", name), slog.Int64("price", price)}
	logVals = append(logVals, logutil.SlogActor(actor)...)
	monitorings.Logger().InfoContext(ctx, "input", logVals...)

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

	p, err := srv.productRepo.Save(ctx, tx, name, price)
	if err != nil {
		return domain.Product{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Product{}, tracerr.Wrap(err)
	}

	return p, nil
}
