package utils

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
)

type Config interface{}

func LoadConfig(cfg Config) error {
	err := godotenv.Load()
	if err != nil {
		log.Println("[Config] No .env file found or error loading .env file. Proceeding with environment variables.")
	}

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("[Config] Error parsing environment variables: %w", err)
	}

	return nil
}
