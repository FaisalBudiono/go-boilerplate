package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"context"
	"errors"
)

type inputRefreshToken interface {
	Context() context.Context
	RefreshToken() string
}

func (srv *Auth) RefreshToken(req inputRefreshToken) (domain.Token, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.auth.refreshToken")
	defer span.End()

	refreshToken := req.RefreshToken()
	payload, err := srv.refreshTokenPayloadParser.ParsePayload(refreshToken)
	if err != nil {
		isInvalidTokenErr := errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenExpired)

		if isInvalidTokenErr {
			return domain.Token{}, errors.Join(ErrInvalidToken, err)
		}

		return domain.Token{}, err
	}

	tx, err := srv.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Token{}, err
	}
	defer tx.Rollback()

	userID, err := srv.authActivityRepo.LastActivityByPayload(ctx, tx, payload)
	if err != nil {
		if errors.Is(err, portout.ErrDataNotFound) {
			return domain.Token{}, errors.Join(ErrInvalidToken, err)
		}

		return domain.Token{}, err
	}

	accessToken, err := srv.jwtUserSigner.Sign(domain.NewUserBasicInfo(domid.UserID(userID)))
	if err != nil {
		return domain.Token{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Token{}, err
	}

	return domain.NewToken(accessToken, refreshToken), nil
}
