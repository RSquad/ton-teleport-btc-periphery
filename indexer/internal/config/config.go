package config

type IndexerConfig struct {
	BitcoinRpcHost            string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser            string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass            string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl              string `env:"COMMON_TON_CONFIG_URL,required"`
	TeleportContractAddr      string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR,required"`
	CoordinatorContractAddr   string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	BitcoinClientContractAddr string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR,required"`
	JettonMinterContractAddr  string `env:"COMMON_TON_CONTRACT_MINTER_ADDR,required"`
	DatabaseURL               string `env:"INDEXER_DATABASE_URL,required"`
}
