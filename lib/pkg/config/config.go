package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
)

type AppConfig interface{}

func LoadConfig[T any]() (T, error) {
	var cfg T

	// Попытка загрузить .env файл, если он существует
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file. Proceeding with environment variables.")
	}

	// Парсинг переменных окружения в структуру конфигурации
	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("ошибка при парсинге переменных окружения: %w", err)
	}

	return cfg, nil
}
