package eagerload

import (
	"context"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
)

func (el *EagerLoad) All(
	ctx context.Context,
	logs []domain.ActivityLog,
) ([]domain.ActivityLogEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(ctx, el.sName("all"))
	defer span.End()

	logIDs := make([]string, len(logs))
	for i, log := range logs {
		logIDs[i] = log.ID
	}
	logIDs = sliceutil.Unique(logIDs)

	logMapMetas, err := el.actLogRepo.GetKeyMapByLogIDs(ctx, el.db, logIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get log metas"),
		)
		return nil, err
	}

	userIDs := make([]string, 0)
	for _, log := range logs {
		if log.ActorID != nil {
			userIDs = append(userIDs, *log.ActorID)
		}
	}
	userIDs = sliceutil.Unique(userIDs)

	userMap, err := el.userRepo.GetMap(ctx, el.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user metas"),
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

	res := make([]domain.ActivityLogEagerLoad, len(logs))
	for i, log := range logs {
		userDetail := func() *domain.UserEagerLoad {
			if log.ActorID == nil {
				return nil
			}
			actorID := *log.ActorID

			user, ok := userMap[actorID]
			if !ok {
				return nil
			}

			roles, ok := userMapRoles[actorID]
			if !ok {
				return nil
			}

			return new(domain.NewUserEagerLoad(user, roles))
		}()

		res[i] = domain.NewActivityLogEagerLoad(
			log,
			logMapMetas[log.ID],
			userDetail,
		)
	}

	return res, nil
}
