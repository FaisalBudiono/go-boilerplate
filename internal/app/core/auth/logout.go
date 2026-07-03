package auth

import (
	"context"
	"errors"
	"log/slog"

	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

type reqLogout interface {
	Context() context.Context
	RefreshToken() string
}

func (srv *Auth) Logout(req reqLogout) error {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("logout"))
	defer span.End()

	refreshToken := req.RefreshToken()

	parsedToken, err := srv.jwtRefreshSigner.Parse(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			monitoring.Logger().DebugContext(
				ctx, "refresh token expired",
				slog.Any("err", err),
			)
			return errors.Join(ErrTokenExpired, err)
		}

		if errs.Is(err, jwt.ErrTokenMalformed, jwt.ErrSignatureInvalid) {
			monitoring.Logger().DebugContext(
				ctx, "refresh token invalid",
				slog.Any("err", err),
			)
			return errors.Join(ErrTokenInvalid, err)
		}

		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to parse refresh token"),
		)
		return err
	}

	tc, err := srv.tokenRepo.FindByClientID(ctx, srv.db, parsedToken.ClientID)
	if err != nil {
		if errors.Is(err, port.ErrDataNotFound) {
			monitoring.Logger().DebugContext(
				ctx, "clientID not found",
				slog.Any("err", err),
			)
			return ErrInvalidCredentials
		}

		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find clientID"),
		)
		return err
	}

	ok, err := srv.hasher.Verify(parsedToken.ClientSecret, tc.HashedSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to verify client secret"),
		)
		return err
	}

	if !ok {
		monitoring.Logger().DebugContext(ctx, "client secret not match")
		return ErrInvalidCredentials
	}

	err = srv.tokenRepo.DeleteByClientID(ctx, srv.db, parsedToken.ClientID)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to delete token credential"),
		)
		return err
	}

	return nil
}
