package req

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func FromCMS(c echo.Context) bool {
	h := c.Request().Header.Get("x-cms-access")

	return strings.ToLower(h) == "true"
}
