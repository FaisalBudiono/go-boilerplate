package app

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

var logLevels = []LogLevel{
	LogLevelDebug,
	LogLevelInfo,
	LogLevelWarn,
	LogLevelError,
}

type envConfig struct {
	AppName string `envconfig:"APP_NAME" default:"go-boilerplate"`

	OtelEndpoint string `envconfig:"OTLP_ENDPOINT" required:"false"`

	PgUser     string `envconfig:"POSTGRES_USER" required:"true"`
	PgPassword string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	PgHost     string `envconfig:"POSTGRES_HOST" required:"true"`
	PgPort     string `envconfig:"POSTGRES_PORT" required:"true"`
	PgDBName   string `envconfig:"POSTGRES_DB_NAME" required:"true"`
	PgSSLMode  string `envconfig:"POSTGRES_SSL_MODE" required:"true"`

	Log struct {
		Level LogLevel `envconfig:"LOG_LEVEL" default:"info" required:"false"`
	}

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

	if !slices.Contains(logLevels, env.Log.Level) {
		validLevels := make([]string, len(logLevels))
		for i, l := range logLevels {
			validLevels[i] = string(l)
		}

		panic(fmt.Sprintf("LOG_LEVEL only support [%s]", strings.Join(validLevels, ",")))
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
