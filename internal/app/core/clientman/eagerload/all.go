package eagerload

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
)

func (el *EagerLoad) All(
	ctx context.Context, ccs []domain.ClientCredential,
) ([]domain.ClientCredentialEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(ctx, el.spanName("all"))
	defer span.End()

	if len(ccs) == 0 {
		return nil, nil
	}

	userIDs := make([]string, len(ccs))
	for i, cc := range ccs {
		userIDs[i] = cc.UserID
	}
	userIDs = sliceutil.Unique(userIDs)

	userMap, err := el.userRepo.GetMap(ctx, el.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user"),
		)
		return nil, err
	}

	userMapRoles, err := el.roleRepo.GetMapByUserID(ctx, el.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user roles"),
		)
		return nil, err
	}

	res := make([]domain.ClientCredentialEagerLoad, len(ccs))
	for i, cc := range ccs {
		u, ok := userMap[cc.UserID]
		if !ok {
			monitoring.Logger().ErrorContext(
				ctx, "user not found", slog.Any("map", userMap),
				slog.String("userID", cc.UserID),
			)
		}

		roles, ok := userMapRoles[cc.UserID]
		if !ok {
			monitoring.Logger().ErrorContext(
				ctx, "user roles not found", slog.Any("map", userMapRoles),
				slog.String("userID", cc.UserID),
			)
		}

		res[i] = domain.NewClientCredentialEagerLoad(
			cc,
			domain.NewUserEagerLoad(u, roles),
		)
	}

	return res, nil
}
