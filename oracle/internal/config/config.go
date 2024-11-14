package config

type OracleConfig struct {
	TonCenterEndpoint          string `env:"TON_CENTER_V2_ENDPOINT,required"`
	Coordinator                string `env:"COORDINATOR,required"`
	Standalone                 bool   `env:"STANDALONE,required"`
	KeystoreDir                string `env:"KEYSTORE_DIR"`
	ValidatorEngineConsolePath string `env:"VALIDATOR_ENGINE_CONSOLE_PATH"`
	ValidatorServerAddress     string `env:"VALIDATOR_SERVER_ADDRESS"`
	ServerPublicKeyPath        string `env:"SERVER_PUBLIC_KEY_PATH"`
	ClientPrivateKeyPath       string `env:"CLIENT_PRIVATE_KEY_PATH"`
}
