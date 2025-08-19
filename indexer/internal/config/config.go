package config

type IndexerConfig struct {
	BitcoinRpcHost          string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser          string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass          string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl            string `env:"COMMON_TON_CONFIG_URL,required"`
	TeleportContractAddr    string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR"`
	CoordinatorContractAddr string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	DatabaseUrl             string `env:"DATABASE_URL,required"`
	DatabaseMaxConn         string `env:"DATABASE_MAX_CONN"`
	DatabaseMaxIdleConn     string `env:"DATABASE_MAX_IDLE_CONN"`
}
