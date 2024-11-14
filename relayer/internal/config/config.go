package config

type RelayerConfig struct {
	BitcoinRpcHost            string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser            string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass            string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl              string `env:"COMMON_TON_CONFIG_URL,required"`
	BitcoinClientContractAddr string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR,required"`
	TeleportContractAddr      string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR,required"`
	RelayerWallerV4Secret     string `env:"RELAYER_WALLET_V4_SECRET,required"`
}
