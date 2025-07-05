package http

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr/authctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr/healthctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr/infoctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr/productctr"
	"FaisalBudiono/go-boilerplate/internal/app/providers"

	"github.com/labstack/echo/v4"
)

func Routes(e *echo.Echo) {
	e.POST("/auths/login", authctr.Login(providers.App().Core.Auth))
	e.POST("/auths/logout", authctr.Logout(providers.App().Core.Auth))
	e.PUT("/auths/refresh", authctr.Refresh(providers.App().Core.Auth))

	e.GET("/health", healthctr.Health(providers.App().Core.Health))

	e.GET("/products", productctr.GetAll(providers.App().Core.Auth, providers.App().Core.Product))
	e.POST("/products", productctr.Save(providers.App().Core.Auth, providers.App().Core.Product))
	e.GET("/products/:productID", productctr.Get(providers.App().Core.Auth, providers.App().Core.Product))
	e.PUT("/products/:productID/publish", productctr.Publish(providers.App().Core.Auth, providers.App().Core.Product))

	e.GET("/userinfo", infoctr.Userinfo(providers.App().Core.Auth))
}
