package config

type IndexerConfig struct {
	BitcoinRpcHost            string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser            string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass            string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl              string `env:"COMMON_TON_CONFIG_URL,required"`
	DatabaseUrl               string `env:"INDEXER_DATABASE_URL,required"`
	RelayerWalletV4Secret     string `env:"RELAYER_WALLET_V4_SECRET,required"`
	TeleportContractAddr      string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR"`
	CoordinatorContractAddr   string `env:"COMMON_TON_CONTRACT_COORDINATOR"`
	BitcoinClientContractAddr string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR"`
	JettonMinterContractAddr  string `env:"COMMON_TON_CONTRACT_MINTER_ADDR"`
	RunServices               string `env:"RUN_SERVICES"`
	Metrics                   string `env:"METRICS"`
	MetricsArgs               string `env:"METRICS_ARGS"`
}
