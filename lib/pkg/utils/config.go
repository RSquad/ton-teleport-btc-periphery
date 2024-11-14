package utils

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func LoadConfig[T any]() (T, error) {
	var cfg T

	if err := godotenv.Load(); err != nil {
		return cfg, err
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
