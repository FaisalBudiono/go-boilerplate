package clientman

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/rnd"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
)

type reqGrantAccess interface {
	Context() context.Context

	UserID() string

	Actor() domain.Userinfo
}

func (srv *ClientManager) GrantAccess(
	req reqGrantAccess,
) (domain.ClientCredentialSecretEagerLoad, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("grant-access"))
	defer span.End()

	userID := req.UserID()
	actor := req.Actor()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.String("userID", userID),
		slog.Any("actor", actor),
	)

	emptyVal := domain.ClientCredentialSecretEagerLoad{}
	if !actor.User.HasRoles(domain.RoleAdmin) {
		return emptyVal, ErrPermissionDenied
	}

	clientSecret, err := func() (string, error) {
		str, err := rnd.String(64)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("secret_%s", str), nil
	}()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to generate clientSecret"),
		)
		return emptyVal, err
	}

	hashedSecret, err := srv.hasher.Hash(clientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to hash clientSecret"),
		)
		return emptyVal, err
	}

	userIDs := []string{userID}
	userMap, err := srv.userRepo.GetMap(ctx, srv.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user target"),
		)
		return emptyVal, err
	}

	userTarget, ok := userMap[userID]
	if !ok {
		monitoring.Logger().InfoContext(
			ctx, "user not found",
			slog.String("userID", userID),
		)
		return emptyVal, ErrUserNotFound
	}

	userMapRoles, err := srv.roleRepo.GetMapByUserID(ctx, srv.db, userIDs)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user roles"),
		)
		return emptyVal, err
	}
	targetRoles, ok := userMapRoles[userID]
	if !ok {
		monitoring.Logger().WarnContext(
			ctx, "target roles not found",
			slog.Any("map", userMapRoles),
		)
		return emptyVal, ErrCannotGrantAdmin
	}

	if slices.Contains(targetRoles, domain.RoleAdmin) {
		monitoring.Logger().DebugContext(ctx, "cannot grant access for admin")
		return emptyVal, ErrCannotGrantAdmin
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin transaction"),
		)

		return emptyVal, err
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

	clientID, err := srv.generateClientID(ctx, tx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to generate clientID"),
		)
		return emptyVal, err
	}

	monitoring.Logger().DebugContext(
		ctx, "inserting client credential",
		slog.String("userID", userTarget.ID),
		slog.String("clientID", clientID),
	)

	err = srv.clientCredRepo.Insert(ctx, tx, domain.NewClientCredentialData(
		userTarget.ID, clientID, hashedSecret,
	))
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert client credential"),
		)
		return emptyVal, err
	}

	clientIDMapCC, err := srv.clientCredRepo.GetMapByClientID(ctx, tx, []string{clientID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find client credential by clientID"),
		)
		return emptyVal, err
	}

	cc, ok := clientIDMapCC[clientID]
	if !ok {
		monitoring.Logger().DebugContext(
			ctx, "clientID not found", slog.String("clientID", clientID),
		)
		return emptyVal, errors.New("clientID not found")
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
	}

	log := srv.activityLogger.New(actlog.ActionGrantAccess)
	log.Actor(&actor)
	log.AddMeta(actlog.MetaKeyClientID, clientID)
	log.AddMeta(actlog.MetaKeyCreatedForID, userID)
	log.AddMeta(actlog.MetaKeyCreatedForName, userTarget.Name)
	err = log.Log(ctx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to log activity"),
		)
		return emptyVal, err
	}

	eagerRes, err := srv.eagerLoad.LoadSecret(ctx, cc, clientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eager load client credential secret"),
		)
		return emptyVal, err
	}

	return eagerRes, nil
}

func (srv *ClientManager) generateClientID(ctx context.Context, tx port.DBTX) (string, error) {
	ctx, span := monitoring.Tracer().Start(ctx, srv.spanName("generate-client-id"))
	defer span.End()

	const maxAttempts = 20
	counter := 0
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			monitoring.Logger().InfoContext(
				ctx, "context cancelled", slog.Any("err", err),
			)
			return "", err
		default:
			if counter > maxAttempts {
				err := errors.New("max attempt generating clientID exceeded")
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to generate clientID"),
				)
				return "", err
			}

			clientID, err := srv.clientID(ctx)
			if err != nil {
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to generate clientID"),
				)
				return "", err
			}

			clientIDMapCC, err := srv.clientCredRepo.GetMapByClientID(ctx, tx, []string{clientID})
			if err != nil {
				otelutil.SpanLogError(
					span, err, otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to find client credential by clientID"),
				)
				return "", err
			}

			_, ok := clientIDMapCC[clientID]
			if !ok {
				return clientID, nil
			}

			counter++
		}
	}
}

func (srv *ClientManager) clientID(ctx context.Context) (string, error) {
	_, span := monitoring.Tracer().Start(ctx, srv.spanName("clientID"))
	defer span.End()

	str, err := rnd.String(32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("id_%s", str), nil
}
