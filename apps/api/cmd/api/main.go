package main

import (
	"context"
	"os"

	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/db"
	apihttp "github.com/bytedance/calandar/apps/api/internal/http"
	"github.com/bytedance/calandar/apps/api/internal/logger"
)

func main() {
	logger.ConfigureFromEnv()
	defer logger.Flush()

	cfg, err := config.Load()
	if err != nil {
		logger.Errorf("api_config_error error=%q", err)
		os.Exit(1)
	}
	logger.Infof(
		"api_starting addr=%s jwt_secret_configured=%t jwks_configured=%t",
		cfg.Addr,
		cfg.SupabaseJWTSecret != "",
		cfg.SupabaseJWKSURL != "",
	)

	conn, err := db.Open(context.Background(), cfg)
	if err != nil {
		logger.Errorf("api_database_connect_failed error=%q", err)
		os.Exit(1)
	}
	logger.Infof("api_database_connected")

	router := apihttp.NewRouter(apihttp.Dependencies{Config: cfg, DB: conn})
	logger.Infof("api_listening addr=%s", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		logger.Errorf("api_server_stopped error=%q", err)
		os.Exit(1)
	}
}
