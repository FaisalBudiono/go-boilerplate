package healthctr

import (
	"context"
	"errors"
	"net/http"

	"FaisalBudiono/go-boilerplate/internal/app/core/healthcheck"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"

	"github.com/labstack/echo/v5"
)

func Health(srv *healthcheck.Healthcheck, version string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, span := monitoring.Tracer().Start(c.Request().Context(), "http.route.health")
		defer span.End()

		deps := srv.Check(&reqHealth{ctx})

		isHealthy := func() bool {
			for _, dep := range deps {
				if !dep.Ok {
					return false
				}
			}
			return true
		}()

		dependencies := func() map[string]string {
			res := make(map[string]string)
			for _, dep := range deps {
				res[dep.Name] = dep.Status
			}

			return res
		}()

		httpStatus := http.StatusOK
		if !isHealthy {
			httpStatus = http.StatusServiceUnavailable

			otelutil.SpanLogError(
				span, errors.New("unhealthy"),
				otelutil.WithErrorLog(ctx),
				otelutil.WithMessage("unhealthy"),
			)
		}

		return c.JSON(httpStatus, healthResponse{
			Healthy:      isHealthy,
			Dependencies: dependencies,
			Version:      version,
		})
	}
}

type reqHealth struct {
	ctx context.Context
}

func (r *reqHealth) Context() context.Context { return r.ctx }

type healthResponse struct {
	Healthy      bool              `json:"healthy"`
	Dependencies map[string]string `json:"dependencies"`

	Version string `json:"version"`
}
