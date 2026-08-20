package config

import (
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"time"
)

type Config struct {
	Addr                 string        `envconfig:"PORT" validate:"required"`
	DatabaseURL          string        `envconfig:"MSSQL_URL" validate:"required"`
	JWTSecret            string        `envconfig:"JWT_SECRET" default:"dev-only-change-me"`
	Environment          string        `envconfig:"ENV" default:"development"`
	TokenTTL             time.Duration `envconfig:"SESSION_TTL" default:"8h"`
	AllowedOrigins       string        `envconfig:"CORS_ORIGINS" validate:"required"`
	RelayTarget          string        `envconfig:"S3RELAY_TARGET" default:"s3relay_ref:5003"`
	HeavenTablePath      string        `envconfig:"HEAVEN_TABLE_PATH" default:"/etc/api-gateway/heaven_table.bin"`
	HeavenServerName     string        `envconfig:"HEAVEN_SERVER_NAME" validate:"required"`
	HeavenServerPassword string        `envconfig:"HEAVEN_SERVER_PASSWORD" validate:"required"`
	HeavenIdentity       string        `envconfig:"HEAVEN_IDENTITY" validate:"required"`
	LogLevel             log.Level     `envconfig:"LOG_LEVEL" default:"info"`
}

var (
	validate = validator.New()
)
var Cfg Config

func initLogConfig() {
	log.SetLevel(Cfg.LogLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func LoadEnv() {
	// Load from .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Parse env vars into the config struct
	if err := envconfig.Process("", &Cfg); err != nil {
		log.Fatalf("Failed to parse env vars: %v", err)
	}

	if err := validate.Struct(Cfg); err != nil {
		log.Fatalf("Configuration validation error: %v", err)
	}
	initLogConfig()
}
