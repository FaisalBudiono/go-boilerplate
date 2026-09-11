package providers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"FaisalBudiono/go-boilerplate/internal/app/adapter/configuration/db"
	"FaisalBudiono/go-boilerplate/internal/app/adapter/out/pg"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth"
	"FaisalBudiono/go-boilerplate/internal/app/core/auth/jwt"
	"FaisalBudiono/go-boilerplate/internal/app/core/healthcheck"
	"FaisalBudiono/go-boilerplate/internal/app/core/user"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/hash"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/mon"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otelutil"
)

type coreConfig struct {
	Auth        *auth.Auth
	Healthcheck *healthcheck.Healthcheck
	User        *user.User
}

type providerConfig struct {
	DB *sql.DB

	Core coreConfig
}

var provider = providerConfig{}

func newShutdowner(sd func() error) *shutdowner {
	return &shutdowner{sd: sd}
}

type shutdowner struct {
	sd func() error
}

func (s *shutdowner) Shutdown() error {
	return s.sd()
}

func Setup(ctx context.Context) ([]*shutdowner, error) {
	ctx, span := mon.Tracer().Start(ctx, "providers.setup")
	defer span.End()

	var shutdowners []*shutdowner
	var err error
	defer func() {
		if err != nil {
			for i, sd := range shutdowners {
				err := sd.Shutdown()
				if err != nil {
					otelutil.SpanLogError(
						span, err, otelutil.WithErrorLog(ctx),
						otelutil.WithMessage(
							fmt.Sprintf("failed to shutdown provider #%d", i),
						),
					)
				}
			}
		}
	}()

	dbconn, err := db.PostgresConn()
	if err != nil {
		return shutdowners, err
	}
	shutdowners = append(shutdowners, newShutdowner(dbconn.Close))

	argonHasher := hash.NewArgon()
	userSigner := jwt.NewUserSigner(
		[]byte(app.ENV().JWT.Secret),
		time.Duration(app.ENV().JWT.TTL)*time.Second,
	)
	refreshSigner := jwt.NewRefreshTokenSigner([]byte(app.ENV().JWT.RefreshSecret))

	userRepo := pg.NewUser()
	tokenRepo := pg.NewTokenCredential()
	roleRepo := pg.NewRole()

	hcCore := healthcheck.New(dbconn)
	authCore := auth.New(
		dbconn,
		userRepo,
		tokenRepo,
		roleRepo,
		argonHasher,
		userSigner,
		refreshSigner,
	)
	userCore := user.New(dbconn, userRepo, roleRepo, argonHasher)

	provider = providerConfig{
		DB: dbconn,

		Core: coreConfig{
			Auth:        authCore,
			Healthcheck: hcCore,
			User:        userCore,
		},
	}

	return shutdowners, nil
}

func App() *providerConfig {
	return &provider
}
