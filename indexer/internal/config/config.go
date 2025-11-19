package config

import (
	"time"
)

type EnvConfig struct {
	BitcoinRpcHost            string        `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser            string        `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass            string        `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl              string        `env:"COMMON_TON_CONFIG_URL,required"`
	CoordinatorContractAddr   string        `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	DatabaseUrl               string        `env:"INDEXER_DATABASE_URL,required"`
	DatabaseMaxConn           string        `env:"INDEXER_DATABASE_MAX_CONN"`
	DatabaseMaxIdleConn       string        `env:"INDEXER_DATABASE_MAX_IDLE_CONN"`
	PProfHttpEnable           string        `env:"INDEXER_PPROF_HTTP_ENABLE"`
	ServerTimeout             time.Duration `env:"LITE_SERVER_TIMEOUT" envDefault:"15s"`
	BitcoinClientContractAddr string        `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR,required"`
	IndexerWalletV4Secret     string        `env:"INDEXER_WALLET_V4_SECRET,required"`
}
