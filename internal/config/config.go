package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

func Load(path string) (Config, error) {
	if err := godotenv.Load(path); err != nil {
		return Config{}, fmt.Errorf("load config file: %w", err)
	}

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8000"
	}

	return cfg, nil
}