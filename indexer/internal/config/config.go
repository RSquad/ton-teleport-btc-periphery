package config

import (
	"fmt"
	"strconv"
)

type IndexerConfig struct {
	BitcoinRpcHost                          string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	BitcoinRpcUser                          string `env:"COMMON_BITCOIN_RPC_USER,required"`
	BitcoinRpcPass                          string `env:"COMMON_BITCOIN_RPC_PASS,required"`
	TonConfigUrl                            string `env:"COMMON_TON_CONFIG_URL,required"`
	TeleportContractAddr                    string `env:"COMMON_TON_CONTRACT_TELEPORT_ADDR"`
	CoordinatorContractAddr                 string `env:"COMMON_TON_CONTRACT_COORDINATOR"`
	BitcoinClientContractAddr               string `env:"COMMON_TON_CONTRACT_BITCLIENT_ADDR"`
	JettonMinterContractAddr                string `env:"COMMON_TON_CONTRACT_MINTER_ADDR"`
	DatabaseURL                             string `env:"INDEXER_DATABASE_URL,required"`
	RunMintService                          string `env:"RUN_MINT_SERVICE"`
	RunEventService                         string `env:"RUN_EVENT_SERVICE"`
	RunPegoutManager                        string `env:"RUN_PEGOUT_MANAGER"`
	RunHttpService                          string `env:"RUN_HTTP_SERVICE"`
	RunMetricsService                       string `env:"RUN_METRICS_SERVICE"`
	RunMetricsFetcherDKG                    string `env:"RUN_METRICS_FETCHER_DKG"`
	RunMetricsFetcherContractBalances       string `env:"RUN_METRICS_FETCHER_CONTRACT_BALANCES"`
	RunMetricsFetcherContractBitcoinClient  string `env:"RUN_METRICS_FETCHER_CONTRACT_BITCOIN_CLIENT"`
	RunMetricsFetcherContractTeleport       string `env:"RUN_METRICS_FETCHER_CONTRACT_TELEPORT"`
	RunMetricsFetcherContractCoordinator    string `env:"RUN_METRICS_FETCHER_CONTRACT_COORDINATOR"`
	RunMetricsFetcherPegouts                string `env:"RUN_METRICS_FETCHER_PEGOUTS"`
	RunMetricsFetcherBitcoinNetwork         string `env:"RUN_METRICS_FETCHER_BITCOIN_NETWORK"`
	MetricsWriterDbChainSize                string `env:"METRICS_WRITE_DB_CHAIN_SIZE"`
	MetricsDkgFetchPeriod                   string `env:"METRICS_DKG_FETCH_PERIOD"`
	MetricsBitcoinClientContractFetchPeriod string `env:"METRICS_BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD"`
	MetricsBitcoinNetworkFetchPeriod        string `env:"METRICS_BITCOIN_NETWORK_FETCH_PERIOD"`
	MetricsTeleportContractFetchPeriod      string `env:"METRICS_TELEPORT_CONTRACT_FETCH_PERIOD"`
	MetricsCoordinatorContractFetchPeriod   string `env:"METRICS_COORDINATOR_CONTRACT_FETCH_PERIOD"`
	MetricsPegoutsFetchPeriod               string `env:"METRICS_PEGOUTS_FETCH_PERIOD"`
}

func ParseIntWithDefaultVal(str string, defaultValue int64, name string) (int64, error) {
	value := defaultValue

	if len(str) > 0 {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", str, name)
		}

		value = val
	}

	return value, nil
}

func ParseBoolWithDefaultVal(str string, defaultValue bool, name string) (bool, error) {
	value := defaultValue

	if len(str) > 0 {
		if str == "true" {
			value = true
		} else if str == "false" {
			value = false
		} else {
			return false, fmt.Errorf("incorrect boolean value '%s' assigned to %s", str, name)
		}
	}

	return value, nil
}
