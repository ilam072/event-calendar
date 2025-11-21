package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"time"
)

type Config struct {
	DB     DBConfig
	Server ServerConfig
	SMTP   SMTPConfig
	JWT    JWTConfig
	Logger LoggerConfig
}

type DBConfig struct {
	PgUser     string `env:"PGUSER"`
	PgPassword string `env:"PGPASSWORD"`
	PgHost     string `env:"PGHOST"`
	PgPort     uint16 `env:"PGPORT"`
	PgDatabase string `env:"PGDATABASE"`
	PgSSLMode  string `env:"PGSSLMODE"`
}

type ServerConfig struct {
	HTTPPort string `env:"HTTP_PORT"`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     string `env:"SMTP_PORT"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
	From     string `env:"FROM"`
}

type JWTConfig struct {
	Secret   string        `env:"SECRET"`
	TokenTTL time.Duration `env:"TOKEN_TTL"`
}

type LoggerConfig struct {
	File string `env:"LOG_FILE"`
}

func MustLoad() *Config {
	cfg := &Config{}

	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
