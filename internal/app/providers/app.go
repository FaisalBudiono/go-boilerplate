package providers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"komdigi-immigration/internal/app/adapter/configuration/db"
	"komdigi-immigration/internal/app/adapter/out/imigserv"
	"komdigi-immigration/internal/app/adapter/out/pg"
	"komdigi-immigration/internal/app/core/auth"
	"komdigi-immigration/internal/app/core/auth/jwt"
	"komdigi-immigration/internal/app/core/clientman"
	"komdigi-immigration/internal/app/core/healthcheck"
	"komdigi-immigration/internal/app/core/imig"
	"komdigi-immigration/internal/app/core/logger"
	"komdigi-immigration/internal/app/core/user"
	"komdigi-immigration/internal/app/core/util/app"
	"komdigi-immigration/internal/app/core/util/hash"
	"komdigi-immigration/internal/app/core/util/monitoring"
	"komdigi-immigration/internal/app/core/util/otelutil"
)

type coreConfig struct {
	ActivityLogger *logger.Logger
	Auth           *auth.Auth
	ClientManager  *clientman.ClientManager
	Healthcheck    *healthcheck.Healthcheck
	Immigration    *imig.Imig
	User           *user.User
}

type providerConfig struct {
	DB *sql.DB

	Core coreConfig
}

var provider = providerConfig{}

type shutdown func() error

func Setup(ctx context.Context) ([]shutdown, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "providers.setup")
	defer span.End()

	var shutdowns []shutdown
	var err error
	defer func() {
		if err != nil {
			for i, sd := range shutdowns {
				err := sd()
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
		return shutdowns, err
	}
	shutdowns = append(shutdowns, dbconn.Close)

	argonHasher := hash.NewArgon()
	userSigner := jwt.NewUserSigner(
		[]byte(app.ENV().JWT.Secret),
		time.Duration(app.ENV().JWT.TTL)*time.Second,
	)
	refreshSigner := jwt.NewRefreshTokenSigner([]byte(app.ENV().JWT.RefreshSecret))

	userRepo := pg.NewUser()
	tokenRepo := pg.NewTokenCredential()
	roleRepo := pg.NewRole()
	clientCredRepo := pg.NewClientCredential()
	activityLogRepo := pg.NewActivityLog()
	immigrationLogRepo := pg.NewImmigrationClearanceLog()

	icService, err := imigserv.New(ctx)
	if err != nil {
		return shutdowns, err
	}

	hcCore := healthcheck.New(dbconn)
	authCore := auth.New(
		dbconn,
		userRepo,
		tokenRepo,
		roleRepo,
		activityLogRepo,
		clientCredRepo,
		argonHasher,
		userSigner,
		refreshSigner,
	)
	userCore := user.New(dbconn, userRepo, roleRepo)
	clientManCore := clientman.New(
		dbconn,
		userRepo,
		roleRepo,
		clientCredRepo,
		activityLogRepo,
		argonHasher,
	)
	activityLogCore := logger.New(
		dbconn,
		activityLogRepo,
		userRepo,
		roleRepo,
	)
	immigrationCore := imig.New(
		dbconn,
		immigrationLogRepo,
		icService,
		activityLogRepo,
		roleRepo,
		userRepo,
	)

	provider = providerConfig{
		DB: dbconn,

		Core: coreConfig{
			ActivityLogger: activityLogCore,
			Auth:           authCore,
			ClientManager:  clientManCore,
			Healthcheck:    hcCore,
			Immigration:    immigrationCore,
			User:           userCore,
		},
	}

	return shutdowns, nil
}

func App() *providerConfig {
	return &provider
}
