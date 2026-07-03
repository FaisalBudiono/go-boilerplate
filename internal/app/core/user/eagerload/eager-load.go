package eagerload

import (
	"context"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
)

func New(
	db port.DBTX,
	roleRepo port.RoleRepo,
) *EagerLoad {
	return &EagerLoad{
		db:       db,
		roleRepo: roleRepo,
	}
}

type EagerLoad struct {
	db port.DBTX

	roleRepo port.RoleRepo
}

func (e *EagerLoad) All(
	ctx context.Context, users []domain.User,
) ([]domain.UserEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "core.user.eagerload.all")
	defer span.End()

	if len(users) == 0 {
		return nil, nil
	}

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	userIDs = sliceutil.Unique(userIDs)

	userIDMapRoles, err := e.roleRepo.GetMapByUserID(ctx, e.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user roles"),
		)
		return nil, err
	}

	res := make([]domain.UserEagerLoad, len(users))
	for i, user := range users {
		roles := userIDMapRoles[user.ID]
		res[i] = domain.NewUserEagerLoad(user, roles)
	}

	return res, nil
}
