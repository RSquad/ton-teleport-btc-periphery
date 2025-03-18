package cfg

type Cfg struct {
	TonConfigUrl               string `env:"COMMON_TON_CONFIG_URL,required"`
	CoordinatorContractAddr    string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	StandaloneMode             bool   `env:"ORACLE_STANDALONE_MODE,required"`
	Pubkey                     string `env:"ORACLE_PUBKEY"`
	Secret                     string `env:"ORACLE_SECRET"`
	ValidatorEngineConsolePath string `env:"ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH"`
	ServerPublicKeyPath        string `env:"ORACLE_SERVER_PUBLIC_KEY_PATH"`
	ClientPrivateKeyPath       string `env:"ORACLE_CLIENT_PRIVATE_KEY_PATH"`
	ValidatorServerAddr        string `env:"ORACLE_VALIDATOR_SERVER_ADDR"`
	KeystorePath               string `env:"ORACLE_KEYSTORE_PATH"`
}
