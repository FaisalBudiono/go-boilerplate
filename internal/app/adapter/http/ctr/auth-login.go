package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res/errcode"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/trace"
)

type reqAuthLogin struct {
	ctx    context.Context
	tracer trace.Tracer

	BodyEmail    string `json:"email" validate:"required"`
	BodyPassword string `json:"password" validate:"required"`
}

func (r *reqAuthLogin) Bind(c echo.Context) error {
	msgs := make(map[string][]string, 0)

	err := c.Bind(r)
	if err != nil {
		var jsonErr *json.UnmarshalTypeError
		if !errors.As(err, &jsonErr) {
			return tracerr.Wrap(err)
		}

		if jsonErr.Field == "email" {
			msgs["email"] = append(msgs["email"], "string")
		}
		if jsonErr.Field == "password" {
			msgs["password"] = append(msgs["password"], "string")
		}
	}

	err = validator.New().StructCtx(r.ctx, r)
	if err != nil {
		var valErr validator.ValidationErrors
		if !errors.As(err, &valErr) {
			return tracerr.Wrap(err)
		}

		for _, fe := range valErr {
			if fe.Field() == "BodyEmail" {
				msgs["email"] = append(msgs["email"], fe.Tag())
			}
			if fe.Field() == "BodyPassword" {
				msgs["password"] = append(msgs["password"], fe.Tag())
			}
		}
	}

	if len(msgs) > 0 {
		return res.NewErrorUnprocessable(msgs)
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

func AuthLogin(tracer trace.Tracer, srv *auth.Auth) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "route: login")
		defer span.End()

		i := &reqAuthLogin{
			ctx:    ctx,
			tracer: tracer,
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

		return c.JSON(http.StatusOK, res.ToAuth(token))
	}
}
