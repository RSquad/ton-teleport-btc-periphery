package config

type OracleConfig struct {
	TonCenterEndpoint       string `env:"COMMON_TON_CENTER_ENDPOINT,required"`
	TonConfigUrl            string `env:"COMMON_TON_CONFIG_URL,required"`
	CoordinatorContractAddr string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
}
