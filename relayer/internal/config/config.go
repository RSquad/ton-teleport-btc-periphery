package config

type RelayerConfig struct {
	BitcoinRpcHost        string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser        string `env:"COMMON_BITCOIN_RPC_USER"`
	BitcoinRpcPass        string `env:"COMMON_BITCOIN_RPC_PASS"`
	TonConfigUrl          string `env:"COMMON_TON_CONFIG_URL,required"`
	ContractBitclientAddr string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR,required"`
	ContractTeleportAddr  string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR,required"`
	WalletSecret          string `env:"RELAYER_WALLET_V4_SECRET,required"`
}
