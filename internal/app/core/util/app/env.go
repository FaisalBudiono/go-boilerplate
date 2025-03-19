package app

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	AppName string `envconfig:"APP_NAME" default:"go-boilerplate"`

	OtelEndpoint string `envconfig:"OTLP_ENDPOINT" required:"false"`

	PgUser     string `envconfig:"POSTGRES_USER" required:"true"`
	PgPassword string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	PgHost     string `envconfig:"POSTGRES_HOST" required:"true"`
	PgPort     string `envconfig:"POSTGRES_PORT" required:"true"`
	PgDBName   string `envconfig:"POSTGRES_DB_NAME" required:"true"`
	PgSSLMode  string `envconfig:"POSTGRES_SSL_MODE" required:"true"`

	JwtSecret        string `envconfig:"JWT_SECRET" required:"true"`
	JwtTTLSecond     int    `envconfig:"JWT_TTL_SECOND" default:"600" required:"false"`
	JwtRefreshSecret string `envconfig:"JWT_REFRESH_SECRET" required:"true"`

	SeederFirstAdminName        string `envconfig:"SEEDER_FIRST_ADMIN_NAME" required:"false" desc:"Name for superadmin (first user)"`
	SeederFirstAdminEmail       string `envconfig:"SEEDER_FIRST_ADMIN_EMAIL" required:"false" desc:"Email for superadmin (first user)"`
	SeederFirstAdminPassword    string `envconfig:"SEEDER_FIRST_ADMIN_PASSWORD" required:"false" desc:"Password for superadmin (first user)"`
	SeederFirstAdminPhoneNumber string `envconfig:"SEEDER_FIRST_ADMIN_PHONE_NUMBER" required:"false" desc:"Phone Number for superadmin (first user)"`
}

var env envConfig

func BindENV() {
	bindDotENV()

	err := envconfig.Process("", &env)
	if err != nil {
		printSpecUsage()
		panic(err)
	}
}

func ENV() envConfig {
	return env
}

func printSpecUsage() {
	err := envconfig.Usage("", &env)
	if err != nil {
		panic(err)
	}
}

func bindDotENV() {
	err := godotenv.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
	}
}
