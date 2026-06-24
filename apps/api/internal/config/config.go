package config

import (
	"errors"
	"os"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	SupabaseJWTSecret string
	SupabaseJWKSURL   string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:              getenv("API_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		SupabaseJWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),
		SupabaseJWKSURL:   os.Getenv("SUPABASE_JWKS_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SupabaseJWTSecret == "" && cfg.SupabaseJWKSURL == "" {
		return Config{}, errors.New("SUPABASE_JWT_SECRET or SUPABASE_JWKS_URL is required")
	}
	return cfg, nil
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
