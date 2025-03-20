package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/trace"
)

type reqPublishProduct struct {
	ctx       context.Context
	actor     domain.User
	productID string
	isPublish bool

	BodyIsPublish *bool `json:"isPublish" validate:"required"`
}

func (r *reqPublishProduct) Bind(c echo.Context) error {
	msgs := make(map[string][]string, 0)

	err := c.Bind(r)
	if err != nil {
		var jsonErr *json.UnmarshalTypeError
		if !errors.As(err, &jsonErr) {
			return tracerr.Wrap(err)
		}

		if jsonErr.Field == "isPublish" {
			msgs["isPublish"] = append(msgs["isPublish"], "boolean")
		}
	}

	err = validator.New().StructCtx(r.ctx, r)
	if err != nil {
		var valErr validator.ValidationErrors
		if !errors.As(err, &valErr) {
			return tracerr.Wrap(err)
		}

		for _, fe := range valErr {
			if fe.Field() == "BodyIsPublish" {
				msgs["isPublish"] = append(msgs["isPublish"], fe.Tag())
			}
		}
	}

	if len(msgs) > 0 {
		return res.NewErrorUnprocessable(msgs)
	}

	r.isPublish = *r.BodyIsPublish

	return nil
}

func (r *reqPublishProduct) Actor() domain.User {
	return r.actor
}

func (r *reqPublishProduct) Context() context.Context {
	return r.ctx
}

func (r *reqPublishProduct) IsPublish() bool {
	return r.isPublish
}

func (r *reqPublishProduct) ProductID() string {
	return r.productID
}

func PublishProduct(tracer trace.Tracer, authSrv *auth.Auth, srv *product.Product) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "route: publish product")
		defer span.End()

		u, err := req.ParseToken(ctx, c, authSrv)
		if err != nil {
			isTokenNotProvidedErr := errors.Is(err, req.ErrInvalidToken) ||
				errors.Is(err, req.ErrNoTokenProvided) ||
				errors.Is(err, req.ErrTokenExpired)

			if isTokenNotProvidedErr {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError(err.Error(), errcode.AuthUnauthorized),
				)
			}
			otel.SpanLogError(span, err, "error when parsing token")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		i := &reqPublishProduct{
			ctx:       ctx,
			productID: c.Param("productID"),
			actor:     u,
		}

		err = i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}
			otel.SpanLogError(span, err, "error when binding request")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		p, err := srv.Publish(i)
		if err != nil {
			if errors.Is(err, product.ErrNotEnoughPermission) {
				return c.JSON(
					http.StatusForbidden,
					res.NewError(err.Error(), errcode.AuthPermissionInsufficient),
				)
			}
			if errors.Is(err, product.ErrNotFound) {
				return c.JSON(
					http.StatusNotFound,
					res.NewError("Product not found", errcode.ProductNotFound),
				)
			}
			otel.SpanLogError(span, err, "error caught in service")

			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusOK, res.ToProduct(p))
	}
}
