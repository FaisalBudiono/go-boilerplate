package authctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqAuthLogin struct {
	ctx context.Context

	BodyEmail    string `json:"email" validate:"required"`
	BodyPassword string `json:"password" validate:"required"`
}

func (r *reqAuthLogin) Bind(c echo.Context) error {
	_, span := monitoring.Tracer().Start(r.ctx, "http.req.auth.login")
	defer span.End()

	errMsgs := make(res.VerboseMetaMsgs, 0)

	validationErr, err := httputil.Bind(c, r, map[string]string{
		"email":    "string",
		"password": "string",
	})
	if err != nil {
		otel.SpanLogError(span, err, "failed to bind")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	validationErr, err = httputil.ValidateStruct(r, map[string]string{
		"BodyEmail":    "email",
		"BodyPassword": "password",
	})
	if err != nil {
		otel.SpanLogError(span, err, "unhandled validator error")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	if len(errMsgs) > 0 {
		return res.NewErrorUnprocessable(errMsgs)
	}

	return nil
}

func (r *reqAuthLogin) Context() context.Context {
	return r.ctx
}

func (r *reqAuthLogin) Email() string {
	return r.BodyEmail
}

func (r *reqAuthLogin) Password() string {
	return r.BodyPassword
}

func Login(srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.ctr.auth.login")
		defer span.End()

		i := &reqAuthLogin{
			ctx: ctx,
		}

		err := i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}

			otel.SpanLogError(span, err, "binding request error")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		token, err := srv.Login(i)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError(err.Error(), errcode.AuthInvalidCredentials),
				)
			}

			otel.SpanLogError(span, err, "error caught in service")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusOK, res.Auth(token))
	}
}
