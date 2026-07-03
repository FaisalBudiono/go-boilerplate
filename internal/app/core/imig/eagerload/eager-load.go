package eagerload

import (
	"context"
	"database/sql"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/sliceutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/port"
)

func New(
	db *sql.DB,
	roleRepo port.RoleRepo,
	userRepo port.UserRepo,
) *EagerLoad {
	return &EagerLoad{
		db:       db,
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

type EagerLoad struct {
	db *sql.DB

	userRepo port.UserRepo
	roleRepo port.RoleRepo
}

func (el *EagerLoad) sName(s string) string {
	return "core.clientman.eagerload." + s
}

func (el *EagerLoad) All(
	ctx context.Context, icLogs []domain.ImmigrationClearanceLog,
) ([]domain.ImmigrationClearanceLogEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(ctx, el.sName("all"))
	defer span.End()

	userIDs := make([]string, 0)
	for _, icl := range icLogs {
		if icl.ActorID != nil {
			userIDs = append(userIDs, *icl.ActorID)
		}
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

	res := make([]domain.ImmigrationClearanceLogEagerLoad, len(icLogs))
	for i, icl := range icLogs {
		actor := func() *domain.UserEagerLoad {
			if icl.ActorID == nil {
				return nil
			}

			u, ok := userMap[*icl.ActorID]
			if !ok {
				monitoring.Logger().ErrorContext(
					ctx, "user not found", slog.Any("map", userMap),
					slog.String("userID", *icl.ActorID),
				)
			}

			roles, ok := userMapRoles[*icl.ActorID]
			if !ok {
				monitoring.Logger().ErrorContext(
					ctx, "user roles not found", slog.Any("map", userMapRoles),
					slog.String("userID", *icl.ActorID),
				)
			}

			return new(domain.NewUserEagerLoad(u, roles))
		}()

		res[i] = domain.NewImmigrationClearanceLogEagerLoad(icl, actor)
	}

	return res, nil
}
