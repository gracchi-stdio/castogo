package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string `env:"DATABASE_URL" envDefault:"postgres://user:password@localhost:5432/mydb?sslmode=disable"`
	SessionSecret string `env:"SESSION_SECRET"`
	Port          string `env:"PORT" envDefault:"8080"`
	Env           string `env:"ENV" envDefault:"development"`
	SiteURL       string `env:"SITE_URL" envDefault:"http://localhost:8080"`
	IsDev         bool
}

var Cfg Config

func LoadConfig() error {
	// Load .env file if it exists (local dev). Docker sets env vars directly.
	godotenv.Load()

	err := env.Parse(&Cfg)
	if err != nil {
		return err
	}

	Cfg.IsDev = Cfg.Env == "development"

	return nil
}
