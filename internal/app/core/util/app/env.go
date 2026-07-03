package app

import (
	"errors"
	"fmt"
	"log"
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
	AppName string `envconfig:"APP_NAME" required:"true"`

	Otel struct {
		LogURL   string `envconfig:"OTLP_LOG_ENDPOINT" required:"false"`
		TraceURL string `envconfig:"OTLP_TRACE_ENDPOINT" required:"false"`
	}

	Log struct {
		Level LogLevel `envconfig:"LOG_LEVEL" default:"info" required:"false"`
	}

	DB struct {
		Postgres struct {
			User     string `envconfig:"POSTGRES_USER" required:"true"`
			Password string `envconfig:"POSTGRES_PASSWORD" required:"true"`
			Host     string `envconfig:"POSTGRES_HOST" required:"true"`
			Port     string `envconfig:"POSTGRES_PORT" required:"true"`
			DBName   string `envconfig:"POSTGRES_DB_NAME" required:"true"`
			SSLMode  string `envconfig:"POSTGRES_SSL_MODE" required:"true"`
		}
	}

	JWT struct {
		Secret        string `envconfig:"JWT_SECRET" required:"true"`
		TTL           int    `envconfig:"JWT_TTL_SECOND" default:"600" required:"true"`
		RefreshSecret string `envconfig:"JWT_REFRESH_SECRET" required:"true"`
	}

	Seeder struct {
		Admin struct {
			Name     string `envconfig:"SEEDER_ADMIN_NAME" required:"true"`
			Email    string `envconfig:"SEEDER_ADMIN_EMAIL" required:"true"`
			Password string `envconfig:"SEEDER_ADMIN_PASSWORD" required:"true"`
		}
	}
}

var env envConfig

func BindENV() {
	bindDotENV()

	err := envconfig.Process("", &env)
	if err != nil {
		printSpecUsage()
		log.Fatalf("failed to process env: %s", err)
	}

	if !slices.Contains(logLevels, env.Log.Level) {
		validLevels := make([]string, len(logLevels))
		for i, l := range logLevels {
			validLevels[i] = string(l)
		}

		log.Fatalf("LOG_LEVEL only support [%s]", strings.Join(validLevels, ","))
	}
}

func ENV() envConfig {
	return env
}

func printSpecUsage() {
	exitCode := 0
	defer func(exit int) {
		if exit != 0 {
			os.Exit(exit)
		}
	}(exitCode)

	err := envconfig.Usage("", &env)
	if err != nil {
		fmt.Println(err)
	}
}

func bindDotENV() {
	err := godotenv.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("failed to load env: %s", err)
		}
	}
}
