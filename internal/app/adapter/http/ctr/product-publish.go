package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqPublishProduct struct {
	ctx context.Context

	actor     domain.User
	productID string
	isPublish bool

	BodyIsPublish *bool `json:"isPublish" validate:"required"`
}

func (r *reqPublishProduct) Bind(c echo.Context) error {
	_, span := monitorings.Tracer().Start(r.ctx, "req: publish product")
	defer span.End()

	errMsgs := make(res.VerboseMetaMsgs, 0)

	validationErr, err := httputil.Bind(c, r, map[string]string{
		"isPublish": "boolean",
	})
	if err != nil {
		otel.SpanLogError(span, err, "failed to bind")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	validationErr, err = httputil.ValidateStruct(r, map[string]string{
		"BodyIsPublish": "isPublish",
	})
	if err != nil {
		otel.SpanLogError(span, err, "unhandled validator error")
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	if len(errMsgs) > 0 {
		return res.NewErrorUnprocessable(errMsgs)
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

func PublishProduct(authSrv *auth.Auth, srv *product.Product) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitorings.Tracer().Start(c.Request().Context(), "route: publish product")
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

		return c.JSON(http.StatusOK, res.Product(p))
	}
}
