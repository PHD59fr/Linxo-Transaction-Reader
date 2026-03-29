package main

import (
	"context"
	"log"

	"linxo-reader/internal/api"
	"linxo-reader/internal/config"
	"linxo-reader/internal/linxo"
	"linxo-reader/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()

	log.Printf("Starting bank API on :%s (debug=%v)", cfg.Port, cfg.Debug)

	fetcher := func(cfg *config.Config) ([]models.Transaction, error) {
		return linxo.FetchTransactions(context.Background(), cfg)
	}

	srv := api.NewServer(cfg, fetcher)
	if err := srv.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
