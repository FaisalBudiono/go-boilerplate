package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/auth"
	"FaisalBudiono/go-boilerplate/internal/app/product"
	"FaisalBudiono/go-boilerplate/internal/domain"
	"FaisalBudiono/go-boilerplate/internal/http/req"
	"FaisalBudiono/go-boilerplate/internal/http/res"
	"FaisalBudiono/go-boilerplate/internal/http/res/errcode"
	"FaisalBudiono/go-boilerplate/internal/otel"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/trace"
)

type reqSaveProduct struct {
	ctx          context.Context
	actor        domain.User
	priceInCents int64

	BodyName  string      `json:"name" validate:"required"`
	BodyPrice json.Number `json:"price"`
}

func (r *reqSaveProduct) Bind(c echo.Context) error {
	msgs := make(map[string][]string, 0)

	err := c.Bind(r)
	if err != nil {
		var jsonErr *json.UnmarshalTypeError
		if !errors.As(err, &jsonErr) {
			return tracerr.Wrap(err)
		}

		if jsonErr.Field == "name" {
			msgs["name"] = append(msgs["name"], "string")
		}
		if jsonErr.Field == "price" {
			msgs["price"] = append(msgs["price"], "number")
		}
	}

	err = validator.New().StructCtx(r.ctx, r)
	if err != nil {
		var valErr validator.ValidationErrors
		if !errors.As(err, &valErr) {
			return tracerr.Wrap(err)
		}

		for _, fe := range valErr {
			if fe.Field() == "BodyName" {
				msgs["name"] = append(msgs["name"], fe.Tag())
			}
		}
	}

	price, err := r.BodyPrice.Int64()
	if err != nil {
		msgs["price"] = append(msgs["price"], "number")
	}
	r.priceInCents = price * 100

	if len(msgs) > 0 {
		return res.NewErrorUnprocessable(msgs)
	}

	return nil
}

func (r *reqSaveProduct) Actor() domain.User {
	return r.actor
}

func (r *reqSaveProduct) Context() context.Context {
	return r.ctx
}

func (r *reqSaveProduct) Name() string {
	return r.BodyName
}

func (r *reqSaveProduct) Price() int64 {
	return r.priceInCents
}

func SaveProduct(tracer trace.Tracer, authSrv *auth.Auth, srv *product.Product) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "route: save product")
		defer span.End()

		u, err := req.ParseToken(ctx, c, authSrv)
		if err != nil {
			isTokenNotProvidedErr := errors.Is(err, req.ErrInvalidToken) ||
				errors.Is(err, req.ErrNoTokenProvided) ||
				errors.Is(err, req.ErrTokenExpired)

			if isTokenNotProvidedErr {
				return c.JSON(http.StatusUnauthorized, res.NewError(err.Error(), errcode.AuthUnauthorized))
			}
			otel.SpanLogError(span, err, "error when parsing token")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		input := &reqSaveProduct{
			ctx:   ctx,
			actor: u,
		}

		err = input.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}
			otel.SpanLogError(span, err, "error when binding request")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		p, err := srv.Save(input)
		if err != nil {
			if errors.Is(err, product.ErrNotEnoughPermission) {
				return c.JSON(
					http.StatusForbidden,
					res.NewError(err.Error(), errcode.AuthPermissionInsufficient),
				)
			}
			if errors.Is(err, product.ErrEmptyName) {
				return c.JSON(
					http.StatusConflict,
					res.NewError(err.Error(), errcode.ProductEmptyName),
				)
			}
			if errors.Is(err, product.ErrNegativePrice) {
				return c.JSON(
					http.StatusConflict,
					res.NewError(err.Error(), errcode.ProductNegativePrice),
				)
			}
			otel.SpanLogError(span, err, "error caught in service")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusCreated, res.ToProduct(p))
	}
}
