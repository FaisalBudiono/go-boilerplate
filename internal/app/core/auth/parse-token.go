package auth

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"context"
	"errors"

	"github.com/ztrue/tracerr"
)

type inputParseToken interface {
	Context() context.Context
	AccessToken() string
}

func (srv *Auth) ParseToken(req inputParseToken) (domain.User, error) {
	ctx, span := monitorings.Tracer().Start(req.Context(), "core.auth.parseToken")
	defer span.End()

	ubasic, err := srv.jwtUserParser.Parse(req.AccessToken())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return domain.User{}, tracerr.Wrap(ErrInvalidToken)
		}

		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return domain.User{}, tracerr.Wrap(ErrInvalidToken)
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.User{}, tracerr.Wrap(ErrTokenExpired)
		}

		return domain.User{}, err
	}

	return srv.userRepo.FindByID(ctx, srv.db, ubasic.ID)
}
