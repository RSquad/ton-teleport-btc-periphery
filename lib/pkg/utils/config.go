package utils

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func LoadCfg[T any]() (T, error) {
	var cfg T

	if err := godotenv.Load(); err != nil {
		log.Printf("[Config] error loading .env file: %v, falling back to default values", err)
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
