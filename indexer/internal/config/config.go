package config

type IndexerConfig struct {
	BitcoinRpcHost       string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser       string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass       string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl         string `env:"COMMON_TON_CONFIG_URL,required"`
	TonCenterV3Host      string `env:"COMMON_TON_CENTER_V3_HOST,required"`
	TonCenterApiKey      string `env:"COMMON_TON_CENTER_API_KEY,required"`
	TeleportContractAddr string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR,required"`
}
