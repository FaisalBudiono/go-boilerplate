package httpfmt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/httpfmt/code/invalid"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
)

// Bind will bind the input using the echo.Context.
//
// fieldTypes is a map of the field name for the request with value of
// the invalid.Code.
//
// e.g. map[string]invalid.Code{"controllerID": invalid.ShouldString}
func Bind(
	c *echo.Context,
	uerr *UnprocessableErr,
	input any,
	fieldTypes map[string]invalid.Code,
) error {
	err := c.Bind(input)
	if err == nil { // err IS NIL
		return nil
	}

	jsonErr, ok := errors.AsType[*json.UnmarshalTypeError](err)
	if !ok {
		return err
	}

	for fieldName, fieldType := range fieldTypes {
		if jsonErr.Field == fieldName {
			uerr.Add(fieldName, fieldType, "")
		}
	}

	return nil
}

func HandleUnprocessable(
	c *echo.Context,
	ctx context.Context,
	span trace.Span,
	err error,
) error {
	if uerr, ok := errors.AsType[*UnprocessableErr](err); ok {
		monitoring.Logger().DebugContext(
			ctx, "unprocessable error",
			slog.Any("error", err),
		)
		return c.JSON(http.StatusUnprocessableEntity, uerr)
	}

	otelutil.SpanLogError(
		span, err, otelutil.WithErrorLog(ctx),
		otelutil.WithMessage("failed to bind request"),
	)
	return c.JSON(
		http.StatusInternalServerError,
		NewErrorGeneric(err, WithTraceID(span)),
	)
}
