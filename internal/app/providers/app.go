package providers

import (
	"FaisalBudiono/go-boilerplate/internal/app/adapter/db"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/pg"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/hash"
	"FaisalBudiono/go-boilerplate/internal/app/core/ht"
	"FaisalBudiono/go-boilerplate/internal/app/core/product"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"database/sql"
	"time"
)

type repoConfig struct {
	AuthActivity *pg.AuthActivity
	Role         *pg.Role
	User         *pg.User
	Product      *pg.Product
}

type coreConfig struct {
	Health  *ht.Healthcheck
	Auth    *auth.Auth
	Product *product.Product
}

type providerConfig struct {
	DB   *sql.DB
	Repo repoConfig
	Core coreConfig
}

var provider = providerConfig{}

func SetUp() {
	dbconn := db.PostgresConn()

	authActivityRepo := pg.NewAuthActivity()
	roleRepo := pg.NewRole()
	userRepo := pg.NewUser(roleRepo)
	productRepo := pg.NewProduct()

	argonHasher := hash.NewArgon()

	jwtUserSigner := jwt.NewUserSigner(
		[]byte(app.ENV().JwtSecret),
		time.Second*time.Duration(app.ENV().JwtTTLSecond),
	)
	refreshTokenSigner := jwt.NewRefreshTokenSigner([]byte(app.ENV().JwtRefreshSecret))

	healthCore := ht.New(dbconn)
	authCore := auth.New(
		dbconn,
		authActivityRepo,
		userRepo,
		argonHasher,
		jwtUserSigner,
		jwtUserSigner,
		refreshTokenSigner,
		refreshTokenSigner,
	)
	productCore := product.New(
		dbconn,
		productRepo,
	)

	provider = providerConfig{
		DB: dbconn,
		Repo: repoConfig{
			AuthActivity: authActivityRepo,
			Role:         roleRepo,
			User:         userRepo,
			Product:      productRepo,
		},
		Core: coreConfig{
			Health:  healthCore,
			Auth:    authCore,
			Product: productCore,
		},
	}
}

func App() *providerConfig {
	return &provider
}
