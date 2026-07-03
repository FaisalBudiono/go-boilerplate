package http

import (
	"komdigi-immigration/internal/app/adapter/in/http/ctr/actlogctr"
	"komdigi-immigration/internal/app/adapter/in/http/ctr/authctr"
	"komdigi-immigration/internal/app/adapter/in/http/ctr/clientcredctr"
	"komdigi-immigration/internal/app/adapter/in/http/ctr/healthctr"
	"komdigi-immigration/internal/app/adapter/in/http/ctr/imigctr"
	"komdigi-immigration/internal/app/adapter/in/http/ctr/userctr"
	"komdigi-immigration/internal/app/core/util/app"
	"komdigi-immigration/internal/app/providers"

	"github.com/labstack/echo/v5"
)

func Routes(e *echo.Echo) {
	authCore := providers.App().Core.Auth

	e.POST("/auth/login", authctr.LoginWithEmail(providers.App().Core.Auth))
	e.POST("/auth/login/client-id", authctr.LoginWithClientID(providers.App().Core.Auth))
	e.PUT("/auth/refresh", authctr.RefreshToken(providers.App().Core.Auth))
	e.POST("/auth/logout", authctr.Logout(providers.App().Core.Auth))

	e.GET("/activity-logs", actlogctr.GetAll(providers.App().Core.ActivityLogger), AuthMiddleware(authCore, WithNeedAuth()))

	e.POST("/immigration/clearance-check", imigctr.ClearanceCheck(providers.App().Core.Immigration), AuthMiddleware(authCore, WithNeedAuth()))
	e.GET("/immigration/imeis/:imei/histories", imigctr.IMEIHistory(providers.App().Core.Immigration), AuthMiddleware(authCore, WithNeedAuth()))

	e.GET("/client-credentials", clientcredctr.GetAll(providers.App().Core.ClientManager), AuthMiddleware(authCore, WithNeedAuth()))
	e.POST("/client-credentials", clientcredctr.GrantAccess(providers.App().Core.ClientManager), AuthMiddleware(authCore, WithNeedAuth()))
	e.DELETE("/client-credentials/:clientID", clientcredctr.RevokeAccess(providers.App().Core.ClientManager), AuthMiddleware(authCore, WithNeedAuth()))

	e.GET("/userinfo", authctr.Userinfo(), AuthMiddleware(authCore, WithNeedAuth()))

	e.GET("/users", userctr.GetAll(providers.App().Core.User), AuthMiddleware(authCore, WithNeedAuth()))

	e.GET("/health", healthctr.Health(providers.App().Core.Healthcheck, app.Version()))
}
