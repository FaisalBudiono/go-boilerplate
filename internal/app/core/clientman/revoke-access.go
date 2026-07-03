package clientman

import (
	"context"
	"errors"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
)

type reqRevokeAccess interface {
	Context() context.Context
	Actor() domain.Userinfo

	ClientID() string
}

func (srv *ClientManager) RevokeAccess(req reqRevokeAccess) error {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("revoke-access"))
	defer span.End()

	actor := req.Actor()
	clientID := req.ClientID()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.String("clientID", clientID),
		slog.Any("actor", actor),
	)

	if !actor.User.HasRoles(domain.RoleAdmin) {
		return ErrPermissionDenied
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin transaction"),
		)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if !errors.Is(err, port.ErrTxDone) {
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to rollback transaction"),
				)
			}
		}
	}()

	clientIDMapCC, err := srv.clientCredRepo.GetMapByClientID(ctx, srv.db, []string{clientID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find client credential by clientID"),
		)
		return err
	}

	cc, ok := clientIDMapCC[clientID]
	if !ok {
		monitoring.Logger().DebugContext(
			ctx, "clientID not found",
			slog.String("clientID", clientID),
		)
		return ErrClientIDNotFound
	}

	userMap, err := srv.userRepo.GetMap(ctx, srv.db, []string{cc.UserID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user by id"),
		)
		return err
	}

	userTarget, ok := userMap[cc.UserID]
	if !ok {
		monitoring.Logger().InfoContext(
			ctx, "user not found", slog.String("userID", cc.UserID),
		)
		return ErrUserNotFound
	}

	err = srv.clientCredRepo.DeleteByClientID(ctx, tx, clientID)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete client credential by clientID"),
		)
		return err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
		return err
	}

	log := srv.activityLogger.New(actlog.ActionRevokeAccess)
	log.Actor(&actor)
	log.AddMeta(actlog.MetaKeyClientID, clientID)
	log.AddMeta(actlog.MetaKeyCreatedForID, userTarget.ID)
	log.AddMeta(actlog.MetaKeyCreatedForName, userTarget.Name)
	err = log.Log(ctx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to log activity"),
		)
		return err
	}

	return nil
}
