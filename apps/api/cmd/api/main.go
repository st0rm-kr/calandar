package main

import (
	"context"
	"log"

	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/db"
	apihttp "github.com/bytedance/calandar/apps/api/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Open(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
	router := apihttp.NewRouter(apihttp.Dependencies{})
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
