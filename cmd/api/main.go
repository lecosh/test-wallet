package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"test-wallet/internal/config"
	"test-wallet/internal/httpapi"
	wallet_api "test-wallet/internal/httpapi/wallet"
	"test-wallet/internal/postgres"
	"test-wallet/internal/wallet"
	"time"
)

// @title Wallet API
// @version 1.0
// @host localhost:8000
// @BasePath /
// @schemes http
func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	if err := run(logger); err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load("config.env")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	cancel()

	if err != nil {
		return err
	}
	defer pool.Close()

	repo := wallet.NewPostgresRepository(pool)
	service := wallet.NewService(repo)
	walletHandler := wallet_api.NewWalletHandler(service)
	walletRouter := wallet_api.NewWalletRouter(walletHandler)

	router := httpapi.NewRouter(walletRouter)
	router.Logger = logger

	if err := router.Start(cfg.HTTPAddr); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start HTTP server: %w", err)
	}

	return nil
}
