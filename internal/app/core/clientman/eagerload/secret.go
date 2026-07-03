package eagerload

import (
	"context"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
)

func (el *EagerLoad) LoadSecret(
	ctx context.Context,
	cc domain.ClientCredential, secret string,
) (domain.ClientCredentialSecretEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(ctx, el.spanName("load-secret"))
	defer span.End()

	emptyVal := domain.ClientCredentialSecretEagerLoad{}

	userIDs := []string{cc.UserID}

	userMap, err := el.userRepo.GetMap(ctx, el.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user"),
		)
		return emptyVal, err
	}

	u, ok := userMap[cc.UserID]
	if !ok {
		monitoring.Logger().ErrorContext(
			ctx, "user not found", slog.Any("map", userMap),
			slog.String("userID", cc.UserID),
		)
		return emptyVal, err
	}

	userMapRoles, err := el.roleRepo.GetMapByUserID(ctx, el.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user roles"),
		)
		return emptyVal, err
	}

	roles, ok := userMapRoles[cc.UserID]
	if !ok {
		monitoring.Logger().ErrorContext(
			ctx, "user roles not found", slog.Any("map", userMapRoles),
			slog.String("userID", cc.UserID),
		)
		return emptyVal, err
	}

	return domain.NewClientCredentialSecretEagerLoad(
		domain.NewClientCredentialSecret(
			cc.ID,
			cc.UserID,
			cc.ClientID,
			secret,
			cc.CreatedAt,
			cc.UpdatedAt,
		),
		domain.NewUserEagerLoad(u, roles),
	), nil
}
