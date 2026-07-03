package http

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/ctr/authctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/ctr/healthctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/in/http/ctr/userctr"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/providers"

	"github.com/labstack/echo/v5"
)

func Routes(e *echo.Echo) {
	authCore := providers.App().Core.Auth

	e.POST("/auth/login", authctr.LoginWithEmail(providers.App().Core.Auth))
	e.PUT("/auth/refresh", authctr.RefreshToken(providers.App().Core.Auth))
	e.POST("/auth/logout", authctr.Logout(providers.App().Core.Auth))

	e.GET("/userinfo", authctr.Userinfo(), AuthMiddleware(authCore, WithNeedAuth()))

	e.POST("/users", userctr.Create(providers.App().Core.User), AuthMiddleware(authCore, WithNeedAuth()))
	e.GET("/users", userctr.GetAll(providers.App().Core.User), AuthMiddleware(authCore, WithNeedAuth()))

	e.GET("/health", healthctr.Health(providers.App().Core.Healthcheck, app.Version()))
}
