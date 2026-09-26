package config

import (
	"os"
)

type Config struct {
	AIStorageKey  string
	Port          string
	AppEnv        string
	AllowedOrigin string
	DatabaseURL   string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://tilik:tilik_secret@localhost:5432/tilik_db?sslmode=disable"
	}

	return &Config{
		AIStorageKey:  os.Getenv("AI_STORAGE_KEY"),
		Port:          port,
		AppEnv:        appEnv,
		AllowedOrigin: allowedOrigin,
		DatabaseURL:   dbURL,
	}
}
