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
	"strconv"

	"github.com/labstack/echo/v4"
)

type reqGetAllProduct struct {
	ctx   context.Context
	actor *domain.User

	page    int64
	perPage int64

	showAllFlag bool
}

func (r *reqGetAllProduct) Bind(c echo.Context) error {
	_, span := monitorings.Tracer().Start(r.ctx, "http.req.product.getAll")
	defer span.End()

	errMsgs := make(res.VerboseMetaMsgs, 0)

	rawPage := c.QueryParam("page")
	if rawPage == "" {
		rawPage = "0"
	}

	integerErr := res.NewVerboseMeta("integer", "")

	page, err := strconv.ParseInt(rawPage, 10, 64)
	if err != nil {
		errMsgs.Append("page", integerErr)
	}

	rawPerPage := c.QueryParam("per_page")
	if rawPerPage == "" {
		rawPerPage = "0"
	}

	perPage, err := strconv.ParseInt(rawPerPage, 10, 64)
	if err != nil {
		errMsgs.Append("per_page", integerErr)
	}

	if len(errMsgs) > 0 {
		return res.NewErrorUnprocessable(errMsgs)
	}

	r.page = page
	r.perPage = perPage
	r.showAllFlag = req.FromCMS(c)

	return nil
}

func (r *reqGetAllProduct) Actor() *domain.User {
	return r.actor
}

func (r *reqGetAllProduct) Context() context.Context {
	return r.ctx
}

func (r *reqGetAllProduct) Page() int64 {
	return r.page
}

func (r *reqGetAllProduct) PerPage() int64 {
	return r.perPage
}

func (r *reqGetAllProduct) CMSAcces() bool {
	return r.showAllFlag
}

func GetAllProduct(
	authSrv *auth.Auth,
	srv *product.Product,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitorings.Tracer().Start(c.Request().Context(), "http.ctr.product.getAll")
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

		i := &reqGetAllProduct{
			ctx:   ctx,
			actor: actor,
		}

		err = i.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.UnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}

			otel.SpanLogError(span, err, "error when binding request")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		products, pg, err := srv.GetAll(i)
		if err != nil {
			otel.SpanLogError(span, err, "error caught in service")
			return c.JSON(http.StatusInternalServerError, res.NewErrorGeneric())
		}

		return c.JSON(http.StatusOK, res.ProductPaginated(products, pg))
	}
}
