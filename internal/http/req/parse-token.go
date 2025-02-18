package req

import (
	"FaisalBudiono/go-boilerplate/internal/app/auth"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"context"
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ztrue/tracerr"
)

type reqParseToken struct {
	ctx   context.Context
	token string
}

func (req *reqParseToken) AccessToken() string {
	return req.token
}

func (req *reqParseToken) Context() context.Context {
	return req.ctx
}

var (
	ErrInvalidToken    = errors.New("Invalid token")
	ErrNoTokenProvided = errors.New("Token not provided")
	ErrTokenExpired    = errors.New("Token expired")
)

func ParseToken(ctx context.Context, c echo.Context, a *auth.Auth) (domain.User, error) {
	tokenHeader := c.Request().Header.Get("authorization")
	if tokenHeader == "" {
		return domain.User{}, tracerr.Wrap(ErrNoTokenProvided)
	}

	tokens := strings.Split(tokenHeader, " ")

	if len(tokens) != 2 {
		return domain.User{}, tracerr.Wrap(ErrInvalidToken)
	}

	u, err := a.ParseToken(&reqParseToken{
		ctx:   ctx,
		token: tokens[1],
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return domain.User{}, tracerr.Wrap(ErrInvalidToken)
		}
		if errors.Is(err, auth.ErrTokenExpired) {
			return domain.User{}, tracerr.Wrap(ErrTokenExpired)
		}

		return domain.User{}, err
	}

	return u, nil
}
