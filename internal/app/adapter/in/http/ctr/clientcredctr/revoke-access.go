package clientcredctr

import (
	"context"
	"net/http"
	"strings"

	"komdigi-immigration/internal/app/adapter/in/http/req"
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

func RevokeAccess(srv *clientman.ClientManager) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(
			c.Request().Context(), "http.route.client-credential.revoke-access",
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

		r := &reqRevokeAccess{ctx: ctx, actor: actor}
		err = r.bind(c)
		if err != nil {
			return httpfmt.HandleUnprocessable(c, ctx, span, err)
		}

		err = srv.RevokeAccess(r)
		if err != nil {
			if errs.Is(err, clientman.ErrPermissionDenied) {
				return c.JSON(http.StatusForbidden, httpfmt.NewError(
					errcode.AuthPermissionDenied,
					"Permission denied",
					httpfmt.WithTraceID(span),
				))
			}

			if errs.Is(err, clientman.ErrClientIDNotFound) {
				return c.JSON(http.StatusNotFound, httpfmt.NewError(
					errcode.NotFound,
					"ClientID not found",
					httpfmt.WithTraceID(span),
				))
			}

			otelutil.SpanLogError(
				span, err, otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("failed revoke access from core"),
			)
			return c.JSON(
				http.StatusInternalServerError,
				httpfmt.NewErrorGeneric(err, httpfmt.WithTraceID(span)),
			)
		}

		return c.NoContent(http.StatusNoContent)
	}
}

type reqRevokeAccess struct {
	ctx   context.Context
	actor *domain.Userinfo

	ParamClientID string `param:"clientID"`

	clientID string
}

func (r *reqRevokeAccess) bind(c *echo.Context) error {
	ctx, span := monitoring.Tracer().Start(
		r.ctx, "http.route.client-credential.revoke-access.bind",
	)
	defer span.End()

	uerr := httpfmt.NewUnprocessableErr(httpfmt.WithTraceID(span))

	err := httpfmt.Bind(c, uerr, r, map[string]invalid.Code{
		"paramClientID": invalid.ShouldString,
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
			"paramClientID", r.ParamClientID,
		): {rules.Required[string]()},
	})
	if err != nil {
		otelutil.SpanLogError(
			span, err, otelutil.WithErrorLog(ctx),
			otelutil.WithMessage("failed to validate request"),
		)
		return err
	}

	clientID := strings.TrimSpace(r.ParamClientID)

	if uerr.IsError() {
		return uerr
	}

	r.clientID = clientID
	return nil
}

func (r *reqRevokeAccess) Actor() domain.Userinfo   { return *r.actor }
func (r *reqRevokeAccess) ClientID() string         { return r.clientID }
func (r *reqRevokeAccess) Context() context.Context { return r.ctx }
