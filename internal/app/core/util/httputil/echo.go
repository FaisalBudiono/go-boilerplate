package httputil

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/errcode"
	"encoding/json"
	"errors"

	"github.com/labstack/echo/v4"
)

// Deprecated: use the Bind
func BindOld(
	c echo.Context, input any, fieldTypes map[string]string,
) (map[string][]domain.OLDVerboseError, error) {
	err := c.Bind(input)
	if err == nil { // err IS NIL
		return nil, nil
	}

	var jsonErr *json.UnmarshalTypeError
	if !errors.As(err, &jsonErr) {
		return nil, err
	}

	m := make(map[string][]domain.OLDVerboseError, 0)

	for fieldName, fieldType := range fieldTypes {
		if jsonErr.Field == fieldName {
			m[fieldName] = append(m[fieldName], domain.OLDNewErrVerbose(errcode.Code(fieldType), ""))
		}
	}

	return m, nil
}

func Bind(
	c echo.Context, input any, fieldTypes map[string]string,
) (map[string][]domain.OLDVerboseError, error) {
	err := c.Bind(input)
	if err == nil { // err IS NIL
		return nil, nil
	}

	var jsonErr *json.UnmarshalTypeError
	if !errors.As(err, &jsonErr) {
		return nil, err
	}

	m := make(map[string][]domain.OLDVerboseError, 0)

	for fieldName, fieldType := range fieldTypes {
		if jsonErr.Field == fieldName {
			m[fieldName] = append(m[fieldName], domain.OLDNewErrVerbose(errcode.Code(fieldType), ""))
		}
	}

	return m, nil
}
