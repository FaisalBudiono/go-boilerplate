package auth

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/rnd"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

type reqLoginWithEmail interface {
	Context() context.Context
	Email() string
	Password() string
}

func (srv *Auth) LoginWithEmail(req reqLoginWithEmail) (domain.TokenPair, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("login-with-email"))
	defer span.End()

	email := req.Email()
	password := req.Password()

	monitoring.Logger().InfoContext(
		ctx, "input",
		slog.String("email", email),
	)

	emptyVal := domain.TokenPair{}

	if email == "" {
		return emptyVal, ErrEmailRequired
	}

	if password == "" {
		return emptyVal, ErrPasswordRequired
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to begin transaction"),
		)

		return emptyVal, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if !errors.Is(err, port.ErrTxDone) {
				otelutil.SpanLogError(
					span, err,
					otelutil.WithErrorLog(ctx),
					otelutil.WithMessage("failed to rollback transaction"),
				)
			}
		}
	}()

	u, err := srv.userRepo.FindByEmail(ctx, tx, email)
	if err != nil {
		if errors.Is(err, port.ErrDataNotFound) {
			monitoring.Logger().DebugContext(ctx, "user not found")
			return emptyVal, errors.Join(ErrInvalidCredentials, err)
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find user by email"),
		)
		return emptyVal, err
	}

	userMapRoles, err := srv.roleRepo.GetMapByUserID(ctx, tx, []string{u.ID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to get roles by user id"),
		)

		return emptyVal, err
	}

	roles, ok := userMapRoles[u.ID]
	if !ok {
		return emptyVal, ErrInvalidCredentials
	}

	if !slices.Contains(roles, domain.RoleAdmin) {
		return emptyVal, ErrDeniedEmailLogin
	}

	match, err := srv.hasher.Verify(password, u.HashedPassword)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to verify password"),
		)

		return emptyVal, err
	}

	if !match {
		monitoring.Logger().WarnContext(ctx, "password mismatch")
		return emptyVal, ErrInvalidCredentials
	}

	tokenClientID := rnd.UUID()
	tokenClientSecret, err := rnd.String(64)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to generate client secret"),
		)

		return emptyVal, err
	}

	tokenSecretHashed, err := srv.hasher.Hash(tokenClientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to hash client secret"),
		)

		return emptyVal, err
	}

	err = srv.tokenRepo.Insert(
		ctx, tx,
		domain.NewTokenCredentialData(
			u.ID,
			tokenClientID,
			tokenSecretHashed,
			domain.LoginMethodEmail,
			u.Email,
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
		domain.NewUserTokenInfo(u.ID, domain.LoginMethodEmail, u.Email),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to sign access token"),
		)

		return emptyVal, err
	}

	refreshToken, err := srv.jwtRefreshSigner.Sign(tokenClientID, tokenClientSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to sign refresh token"),
		)

		return emptyVal, err
	}

	err = tx.Commit()
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to commit transaction"),
		)

		return emptyVal, err
	}

	return domain.NewTokenPair(accessToken, refreshToken), nil
}
