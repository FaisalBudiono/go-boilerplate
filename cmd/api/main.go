package main

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/db"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/http/ctr"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/otel"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/pg"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/hash"
	"FaisalBudiono/go-boilerplate/internal/app/core/ht"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func main() {
	app.BindENV()

	ctx := context.Background()

	shutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			panic(err)
		}
	}()

	tracer := otel.NewTracer(app.ENV().AppName)
	logger := otel.NewLogger(app.ENV().AppName)

	dbconn := db.PostgresConn()

	authActivityRepo := pg.NewAuthActivity(tracer)
	roleRepo := pg.NewRole(tracer)
	userRepo := pg.NewUser(tracer, roleRepo)
	productRepo := pg.NewProduct(tracer)

	argonHasher := hash.NewArgon()

	jwtUserSigner := jwt.NewUserSigner(
		[]byte(app.ENV().JwtSecret),
		time.Second*time.Duration(app.ENV().JwtTTLSecond),
	)
	refreshTokenSigner := jwt.NewRefreshTokenSigner([]byte(app.ENV().JwtRefreshSecret))

	healthSrv := ht.New(dbconn, tracer, logger)

	authSrv := auth.New(
		dbconn,
		tracer,
		authActivityRepo,
		userRepo,
		argonHasher,
		jwtUserSigner,
		jwtUserSigner,
		refreshTokenSigner,
		refreshTokenSigner,
	)

	productSrv := product.New(
		dbconn,
		tracer,
		productRepo,
	)

	e := echo.New()
	e.Use(otelecho.Middleware(app.ENV().AppName))

	e.POST("/auths/login", ctr.AuthLogin(tracer, authSrv))
	e.POST("/auths/logout", ctr.AuthLogout(tracer, authSrv))
	e.PUT("/auths/refresh", ctr.AuthRefresh(tracer, authSrv))

	e.GET("/health", ctr.Health(tracer, logger, healthSrv))

	e.GET("/products", ctr.GetAllProduct(tracer, authSrv, productSrv))
	e.POST("/products", ctr.SaveProduct(tracer, authSrv, productSrv))
	e.GET("/products/:productID", ctr.GetProduct(tracer, authSrv, productSrv))
	e.PUT("/products/:productID/publish", ctr.PublishProduct(tracer, authSrv, productSrv))

	e.GET("/userinfo", ctr.Userinfo(tracer, authSrv))

	e.Logger.Fatal(e.Start(":8080"))
}
