package config

type EnvConfig struct {
	BitcoinRpcHost                   string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser                   string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass                   string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl                     string `env:"COMMON_TON_CONFIG_URL,required"`
	TeleportContractAddr             string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR,required"`
	CoordinatorContractAddr          string `env:"COMMON_TON_CONTRACT_COORDINATOR,required"`
	BitcoinClientContractAddr        string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR,required"`
	JettonMinterContractAddr         string `env:"COMMON_TON_CONTRACT_MINTER_ADDR,required"`
	RelayerWalletAddr                string `env:"COMMON_TON_RELAYER_WALLET_ADDR,required"`
	IndexerWalletAddr                string `env:"COMMON_TON_INDEXER_WALLET_ADDR,required"`
	DatabaseUrl                      string `env:"METRICS_DATABASE_URL,required"`
	DatabaseMaxConn                  string `env:"METRICS_DATABASE_MAX_CONN"`
	DatabaseMaxIdleConn              string `env:"METRICS_DATABASE_MAX_IDLE_CONN"`
	HttpPort                         string `env:"METRICS_HTTP_PORT"`
	WriterDbChainSize                string `env:"WRITE_DB_CHAIN_SIZE"`
	DkgFetchPeriod                   string `env:"DKG_FETCH_PERIOD"`
	BitcoinClientContractFetchPeriod string `env:"BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD"`
	BitcoinNetworkFetchPeriod        string `env:"BITCOIN_NETWORK_FETCH_PERIOD"`
	TeleportContractFetchPeriod      string `env:"TELEPORT_CONTRACT_FETCH_PERIOD"`
	CoordinatorContractFetchPeriod   string `env:"COORDINATOR_CONTRACT_FETCH_PERIOD"`
	ContractBalancesFetchPeriod      string `env:"CONTRACT_BALANCES_FETCH_PERIOD"`
	PProfHttpEnable                  string `env:"METRICS_PPROF_HTTP_ENABLE"`
	TonExplorer                      string `env:"TON_EXPLORER"`
	BtcExplorer                      string `env:"BTC_EXPLORER"`
	Runbook                          string `env:"RUNBOOK"`
	AlertsTestApiEnable              string `env:"ALERTS_TEST_API_ENABLE"`
	AlertsCheckPeriod                string `env:"ALERTS_CHECK_PERIOD"`
	AlertBtcBlockDeltaHeightWarn     string `env:"ALERT_BTC_BLOCK_DELTA_HEIGHT_WARN"`
	AlertBtcBlockDeltaHeightCrit     string `env:"ALERT_BTC_BLOCK_DELTA_HEIGHT_CRIT"`
	AlertCooldownPeriod              string `env:"ALERT_COOLDOWN_PERIOD"`
	TgBotToken                       string `env:"METRICS_TELEGRAM_BOT_TOKEN"`
	TgChatId                         string `env:"METRICS_TELEGRAM_CHAT_ID"`
}
