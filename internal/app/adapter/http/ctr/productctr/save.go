package productctr

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/req"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/res"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/httputil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqSaveProduct struct {
	ctx          context.Context
	actor        domain.User
	priceInCents int64

	BodyName  string `json:"name" validate:"required"`
	BodyPrice int64  `json:"price"`
}

func (r *reqSaveProduct) Bind(c echo.Context) error {
	ctx, span := monitoring.Tracer().Start(r.ctx, "http.req.product.save")
	defer span.End()

	errMsgs := make(res.OLDVerboseMetaMsgs, 0)

	validationErr, err := httputil.BindOld(c, r, map[string]string{
		"name":  "string",
		"price": "integer",
	})
	if err != nil {
		otel.SpanLogError(span, err,
			otel.WithErrorLog(ctx),
			otel.WithMessage("failed to bind"),
		)
		return err
	}
	errMsgs.AppendDomMap(validationErr)

	validationErr, err = httputil.ValidateStruct(r, map[string]string{
		"BodyName":  "name",
		"BodyPrice": "price",
	})
	if err != nil {
		otel.SpanLogError(span, err,
			otel.WithErrorLog(ctx),
			otel.WithMessage("unhandled validator error"),
		)

		return err
	}
	errMsgs.AppendDomMap(validationErr)

	if len(errMsgs) > 0 {
		return res.OLDNewErrorUnprocessable(errMsgs)
	}

	r.priceInCents = r.BodyPrice * 100

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

func Save(authSrv *auth.Auth, srv *product.Product) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.ctr.product.save")
		defer span.End()

		u, err := req.ParseToken(ctx, c, authSrv)
		if err != nil {
			isTokenNotProvidedErr := errors.Is(err, req.ErrInvalidToken) ||
				errors.Is(err, req.ErrNoTokenProvided) ||
				errors.Is(err, req.ErrTokenExpired)

			if isTokenNotProvidedErr {
				return c.JSON(http.StatusUnauthorized, res.OLDNewError(err.Error(), errcode.AuthUnauthorized))
			}

			otel.SpanLogError(span, err,
				otel.WithErrorLog(ctx),
				otel.WithMessage("error when parsing token"),
			)
			return c.JSON(http.StatusInternalServerError, res.OLDNewErrorGeneric())
		}

		input := &reqSaveProduct{
			ctx:   ctx,
			actor: u,
		}

		err = input.Bind(c)
		if err != nil {
			if unErr, ok := err.(*res.OLDUnprocessableErrResponse); ok {
				return c.JSON(http.StatusUnprocessableEntity, unErr)
			}

			otel.SpanLogError(span, err,
				otel.WithErrorLog(ctx),
				otel.WithMessage("error when binding request"),
			)
			return c.JSON(http.StatusInternalServerError, res.OLDNewErrorGeneric())
		}

		p, err := srv.Save(input)
		if err != nil {
			if errors.Is(err, product.ErrNotEnoughPermission) {
				return c.JSON(
					http.StatusForbidden,
					res.OLDNewError(err.Error(), errcode.AuthPermissionInsufficient),
				)
			}
			if errors.Is(err, product.ErrEmptyName) {
				return c.JSON(
					http.StatusConflict,
					res.OLDNewError(err.Error(), errcode.ProductEmptyName),
				)
			}
			if errors.Is(err, product.ErrNegativePrice) {
				return c.JSON(
					http.StatusConflict,
					res.OLDNewError(err.Error(), errcode.ProductNegativePrice),
				)
			}

			otel.SpanLogError(span, err,
				otel.WithErrorLog(ctx),
				otel.WithMessage("error caught in service"),
			)
			return c.JSON(http.StatusInternalServerError, res.OLDNewErrorGeneric())
		}

		return c.JSON(http.StatusCreated, res.Product(p))
	}
}
