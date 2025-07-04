package http

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr"
	"FaisalBudiono/go-boilerplate/internal/app/providers"

	"github.com/labstack/echo/v4"
)

func Routes(e *echo.Echo) {
	e.POST("/auths/login", ctr.AuthLogin(providers.App().Core.Auth))
	e.POST("/auths/logout", ctr.AuthLogout(providers.App().Core.Auth))
	e.PUT("/auths/refresh", ctr.AuthRefresh(providers.App().Core.Auth))

	e.GET("/health", ctr.Health(providers.App().Core.Health))

	e.GET("/products", ctr.GetAllProduct(providers.App().Core.Auth, providers.App().Core.Product))
	e.POST("/products", ctr.SaveProduct(providers.App().Core.Auth, providers.App().Core.Product))
	e.GET("/products/:productID", ctr.GetProduct(providers.App().Core.Auth, providers.App().Core.Product))
	e.PUT("/products/:productID/publish", ctr.PublishProduct(providers.App().Core.Auth, providers.App().Core.Product))

	e.GET("/userinfo", ctr.Userinfo(providers.App().Core.Auth))
}
