package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, falling back to default values")
	}

	checkEnv("COMMON_BITCOIN_RPC_HOST")
	checkEnv("COMMON_BITCOIN_RPC_USER")
	checkEnv("COMMON_BITCOIN_RPC_PASS")
	checkEnv("COMMON_TON_CONFIG_URL")
	checkEnv("COMMON_TON_CONTRACT_BITCLIENT_ADDR")
	checkEnv("COMMON_TON_CONTRACT_TELEPORT_ADDR")
	checkEnv("RELAYER_WALLET_V4_SECRET")
}

func checkEnv(name string) {
	env := os.Getenv(name)
	if env == "" {
		err := fmt.Errorf("environment variable %s is not set", name)
		panic(err)
	}
}
