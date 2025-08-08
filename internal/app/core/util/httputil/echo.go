package httputil

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"encoding/json"
	"errors"

	"github.com/labstack/echo/v4"
)

func Bind(
	c echo.Context, input any, fieldTypes map[string]string,
) (map[string][]domain.VerboseError, error) {
	err := c.Bind(input)
	if err == nil {
		return nil, nil
	}

	var jsonErr *json.UnmarshalTypeError
	if !errors.As(err, &jsonErr) {
		return nil, err
	}

	m := make(map[string][]domain.VerboseError, 0)

	for fieldName, fieldType := range fieldTypes {
		if jsonErr.Field == fieldName {
			m[fieldName] = append(m[fieldName], domain.NewErrVerbose(errcode.Code(fieldType), ""))
		}
	}

	return m, nil
}
