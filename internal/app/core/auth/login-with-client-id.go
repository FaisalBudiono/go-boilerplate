package auth

import (
	"context"
	"errors"
	"log/slog"

	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/core/util/rnd"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/actlog"
	"komdigi-immigration/internal/app/port"
)

type reqLoginWithClientID interface {
	Context() context.Context

	ClientID() string
	ClientSecret() string
}

func (srv *Auth) LoginWithClientID(
	req reqLoginWithClientID,
) (domain.TokenPair, error) {
	ctx, span := monitoring.Tracer().Start(
		req.Context(), srv.spanName("login-with-client-id"),
	)
	defer span.End()

	clientID := req.ClientID()
	clientSecret := req.ClientSecret()

	monitoring.Logger().InfoContext(
		ctx, "input", slog.String("clientID", clientID),
	)

	emptyVal := domain.TokenPair{}
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

	cc, err := srv.clientCredRepo.FindSecretByClientID(ctx, tx, clientID)
	if err != nil {
		if errors.Is(err, port.ErrDataNotFound) {
			monitoring.Logger().DebugContext(
				ctx, "clientID not found",
				slog.Any("err", err),
			)
			return emptyVal, ErrInvalidCredentials
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find clientID"),
		)
		return emptyVal, err
	}

	match, err := srv.hasher.Verify(clientSecret, cc.Secret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to verify clientSecret"),
		)
		return emptyVal, err
	}

	if !match {
		monitoring.Logger().WarnContext(ctx, "client secret mismatch")
		return emptyVal, ErrInvalidCredentials
	}

	tokenClientID := rnd.UUID()
	tokenClientSecret, err := rnd.String(64)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to generate token client secret"),
		)

		return emptyVal, err
	}

	tokenSecretHashed, err := srv.hasher.Hash(tokenClientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to hash token client secret"),
		)

		return emptyVal, err
	}

	err = srv.tokenRepo.Insert(
		ctx, tx,
		domain.NewTokenCredentialData(
			cc.CC.UserID,
			tokenClientID,
			tokenSecretHashed,
			domain.LoginMethodClientID,
			cc.CC.ClientID,
		),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to insert token to repo"),
		)

		return emptyVal, err
	}

	accessToken, err := srv.jwtUserSigner.Sign(
		domain.NewUserTokenInfo(cc.CC.UserID, domain.LoginMethodClientID, cc.CC.ClientID),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to sign access token"),
		)
		return emptyVal, err
	}

	refreshToken, err := srv.jwtRefreshSigner.Sign(tokenClientID, tokenClientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to sign refresh token"),
		)
		return emptyVal, err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)
		return emptyVal, err
	}

	userMap, err := srv.userRepo.GetMap(ctx, srv.db, []string{cc.CC.UserID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get user by id"),
		)
		return emptyVal, err
	}

	actorUser, ok := userMap[cc.CC.UserID]
	if !ok {
		monitoring.Logger().ErrorContext(ctx, "user not found")
		return emptyVal, ErrUserNotFound
	}

	eagerUsers, err := srv.userLoader.All(ctx, []domain.User{actorUser})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get eager user"),
		)
		return emptyVal, err
	}

	eagerActor := eagerUsers[0]

	actor := domain.NewUserinfo(
		eagerActor,
		domain.NewUserTokenInfo(
			cc.CC.UserID, domain.LoginMethodClientID, cc.CC.ClientID,
		),
	)

	log := srv.activityLogger.New(actlog.ActionLogin)

	log.Actor(&actor)
	err = log.Log(ctx)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to log activity"),
		)
		return emptyVal, err
	}

	return domain.NewTokenPair(accessToken, refreshToken), nil
}
