package config

import (
	"errors"
	"os"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	SupabaseJWTSecret string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:              getenv("API_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		SupabaseJWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SupabaseJWTSecret == "" {
		return Config{}, errors.New("SUPABASE_JWT_SECRET is required")
	}
	return cfg, nil
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
