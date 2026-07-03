package clientcredctr

import (
	"context"
	"net/http"
	"strings"

	"komdigi-immigration/internal/app/adapter/in/http/req"
	"komdigi-immigration/internal/app/adapter/in/http/res"
	"komdigi-immigration/internal/app/core/clientman"
	"komdigi-immigration/internal/app/core/util/errs"
	"komdigi-immigration/internal/app/core/util/httpfmt"
	"komdigi-immigration/internal/app/core/util/httpfmt/code/invalid"
	"komdigi-immigration/internal/app/core/util/httpfmt/rules"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
	"komdigi-immigration/internal/app/domain"
	"komdigi-immigration/internal/app/domain/errcode"

	"github.com/labstack/echo/v5"
)

func GrantAccess(srv *clientman.ClientManager) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.client-credential.grant-access",
		)
		defer span.End()

		actor, err := req.AuthUser(ctx)
		if err != nil {
			monitoring.Logger().DebugContext(
				ctx, "failed to get authenticated user",
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		r := &reqGrantAccess{ctx: ctx, actor: *actor}

		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		cc, err := srv.GrantAccess(r)
		if err != nil {
			if errs.Is(err, clientman.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, clientman.ErrUserNotFound) {
				return c.JSON(http.StatusNotFound, httpfmt.NewError(
					errcode.NotFound,
					"User not found",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, clientman.ErrCannotGrantAdmin) {
				return c.JSON(http.StatusConflict, httpfmt.NewError(
					errcode.ClientCredFailedGrant,
					"Cannot grant access for admin",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed grant access from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.JSON(http.StatusCreated, res.ClientCredSecret(cc))
	}
}

type reqGrantAccess struct {
	ctx   context.Context
	actor domain.Userinfo

	BodyUserID string `json:"userID"`

	userID string
}

func (r *reqGrantAccess) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.client-credential.grant-access.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"userID": invalid.ShouldString,
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to bind request"),
		)
		return err
	}

	err = httpfmt.Validate(ctx, uerr, map[httpfmt.ValidationInput][]httpfmt.Rule{
		httpfmt.NewValidationInput(
			"userID", r.BodyUserID,
		): {rules.Required[string]()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	userID := strings.TrimSpace(r.BodyUserID)

	if uerr.IsError() {
		return uerr
	}

	r.userID = userID

	return nil
}

func (r *reqGrantAccess) Context() context.Context { return r.ctx }
func (r *reqGrantAccess) Actor() domain.Userinfo   { return r.actor }
func (r *reqGrantAccess) UserID() string           { return r.userID }
