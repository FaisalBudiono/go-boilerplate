package ctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqGetProduct struct {
	ctx       context.Context
	actor     *domain.User
	productID string
}

func (r *reqGetProduct) Actor() *domain.User {
	return r.actor
}

func (r *reqGetProduct) Context() context.Context {
	return r.ctx
}

func (r *reqGetProduct) ProductID() string {
	return r.productID
}

func GetProduct(
	authSrv *auth.Auth,
	srv *product.Product,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitorings.Tracer().Start(c.Request().Context(), "route: get product")
		defer span.End()

		u, err := req.ParseToken(ctx, c, authSrv)
		if err != nil {
			isTokenNotProvidedErr := errors.Is(err, req.ErrInvalidToken) ||
				errors.Is(err, req.ErrTokenExpired)

			if isTokenNotProvidedErr {
				return c.JSON(
					http.StatusUnauthorized,
					res.NewError(err.Error(), errcode.AuthUnauthorized),
				)
			}
			if !errors.Is(err, req.ErrNoTokenProvided) {
				otel.SpanLogError(span, err, "error when parsing token")
				return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
			}
		}

		var actor *domain.User
		if err == nil {
			actor = &u
		}

		i := &reqGetProduct{
			ctx:       ctx,
			actor:     actor,
			productID: c.Param("productID"),
		}

		p, err := srv.Get(i)
		if err != nil {
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
