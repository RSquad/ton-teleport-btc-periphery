package config

type OracleConfig struct {
	TonCenterEndpoint          string `env:"COMMON_TON_CENTER_ENDPOINT,required"`
	CoordinatorContractAddr    string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	TonConfigUrl               string `env:"COMMON_TON_CONFIG_URL,required"`
	IsStandalone               bool   `env:"ORACLE_IS_STANDALONE,required"`
	KeystoreDir                string `env:"ORACLE_KEYSTORE_DIR"`
	ValidatorEngineConsolePath string `env:"ORACLE_VALIDATOR_ENGINE_CONSOLE_PATH"`
	ValidatorServerAddress     string `env:"ORACLE_VALIDATOR_SERVER_ADDRESS"`
	ServerPublicKeyPath        string `env:"ORACLE_SERVER_PUBLIC_KEY_PATH"`
	ClientPrivateKeyPath       string `env:"ORACLE_CLIENT_PRIVATE_KEY_PATH"`
}
