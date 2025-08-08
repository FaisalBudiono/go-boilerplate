package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"errors"
)

type inputParseToken interface {
	Context() context.Context
	AccessToken() string
}

func (srv *Auth) ParseToken(req inputParseToken) (domain.User, error) {
	ctx, span := monitoring.Tracer().Start(req.Context(), "core.auth.parseToken")
	defer span.End()

	ubasic, err := srv.jwtUserParser.Parse(req.AccessToken())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return domain.User{}, errors.Join(ErrInvalidToken, err)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return domain.User{}, errors.Join(ErrInvalidToken, err)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.User{}, errors.Join(ErrTokenExpired, err)
		}

		return domain.User{}, err
	}

	return srv.userRepo.FindByID(ctx, srv.db, ubasic.ID)
}
