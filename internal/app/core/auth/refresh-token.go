package auth

import (
	"context"
	"errors"
	"log/slog"

	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/errs"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port"
)

type reqRefreshToken interface {
	Context() context.Context
	RefreshToken() string
}

func (srv *Auth) RefreshToken(req reqRefreshToken) (domain.TokenPair, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("refresh-token"))
	defer span.End()

	refreshToken := req.RefreshToken()

	emptyVal := domain.TokenPair{}

	parsedToken, err := srv.jwtRefreshSigner.Parse(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			monitoring.Logger().DebugContext(
				ctx, "refresh token expired",
				slog.Any("err", err),
			)

			return emptyVal, errors.Join(ErrTokenExpired, err)
		}

		if errs.Is(err, jwt.ErrTokenMalformed, jwt.ErrSignatureInvalid) {
			monitoring.Logger().DebugContext(
				ctx, "refresh token invalid",
				slog.Any("err", err),
			)

			return emptyVal, errors.Join(ErrTokenInvalid, err)
		}

		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to parse refresh token"),
		)

		return emptyVal, err
	}

	tc, err := srv.tokenRepo.FindByClientID(ctx, srv.db, parsedToken.ClientID)
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

	ok, err := srv.hasher.Verify(parsedToken.ClientSecret, tc.HashedSecret)
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to verify client secret"),
		)
		return emptyVal, err
	}

	if !ok {
		monitoring.Logger().DebugContext(ctx, "client secret not match")
		return emptyVal, ErrInvalidCredentials
	}

	accessToken, err := srv.jwtUserSigner.Sign(
		domain.NewUserTokenInfo(tc.UserID, tc.LoginMethod, tc.LoginID),
	)
	if err != nil {
		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to sign access token"),
		)

		return emptyVal, err
	}

	return domain.NewTokenPair(accessToken, refreshToken), nil
}
