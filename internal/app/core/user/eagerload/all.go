package eagerload

import (
	"context"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/sliceutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
)

func (el *EagerLoad) All(
	ctx context.Context, users []domain.User,
) ([]domain.UserEagerLoad, error) {
	ctx, span := mon.Tracer().Start(ctx, el.sName("all"))
	defer span.End()

	if len(users) == 0 {
		return nil, nil
	}

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	userIDs = sliceutil.Unique(userIDs)

	userIDMapRoles, err := el.roleRepo.GetMapByUserID(ctx, el.db, userIDs)
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
