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
)

type reqParseUser interface {
	Context() context.Context
	AccessToken() string
}

func (srv *Auth) ParseUser(req reqParseUser) (domain.Userinfo, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), srv.spanName("parse-user"))
	defer span.End()

	accessToken := req.AccessToken()

	emptyVal := domain.Userinfo{}

	uTokenInfo, err := srv.jwtUserSigner.Parse(accessToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			monitoring.Logger().DebugContext(
				ctx, "access token expired",
				slog.Any("err", err),
			)
			return emptyVal, errors.Join(ErrTokenExpired, err)
		}

		if errs.Is(err, jwt.ErrTokenMalformed, jwt.ErrSignatureInvalid) {
			monitoring.Logger().DebugContext(
				ctx, "access token invalid",
				slog.Any("err", err),
			)
			return emptyVal, errors.Join(ErrTokenInvalid, err)
		}

		otelutil.SpanLogError(
			span, err,
			otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to parse access token"),
		)
		return emptyVal, err
	}

	userMap, err := srv.userRepo.GetMap(ctx, srv.db, []string{uTokenInfo.ID})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to find user by ID"),
		)
		return emptyVal, err
	}
	u, ok := userMap[uTokenInfo.ID]
	if !ok {
		monitoring.Logger().WarnContext(
			ctx, "user not found",
			slog.String("userID", uTokenInfo.ID),
		)
		return emptyVal, ErrUserNotFound
	}

	res, err := srv.userLoader.All(ctx, []domain.User{u})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to eagerload user"),
		)
		return emptyVal, err
	}

	return domain.NewUserinfo(res[0], uTokenInfo), nil
}
