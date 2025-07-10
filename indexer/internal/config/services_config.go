package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xssnick/tonutils-go/address"
)

type MetricsConfig struct {
	RunFetcherDKG                   bool
	RunFetcherContractBalances      bool
	RunFetcherContractBitcoinClient bool
	RunFetcherContractTeleport      bool
	RunFetcherContractCoordinator   bool
	RunFetcherPegouts               bool
	RunFetcherBitcoinNetwork        bool

	WriterDbChainSize                int
	DkgFetchPeriod                   int
	BitcoinClientContractFetchPeriod int
	BitcoinNetworkFetchPeriod        int
	TeleportContractFetchPeriod      int
	CoordinatorContractFetchPeriod   int
}

type RunServicesConfig struct {
	RunMintService    bool
	RunEventService   bool
	RunPegoutManager  bool
	RunHttpService    bool
	RunMetricsService bool
}

type ExternalServicesConfig struct {
	BitcoinRpcHost            string
	BitcoinRpcUser            string
	BitcoinRpcPass            string
	TonConfigUrl              string
	DatabaseUrl               string
	TeleportContractAddr      *address.Address
	CoordinatorContractAddr   *address.Address
	BitcoinClientContractAddr *address.Address
	JettonMinterContractAddr  *address.Address
}

type ServicesConfig struct {
	ExternalServices *ExternalServicesConfig
	RunServices      *RunServicesConfig
	Metrics          *MetricsConfig
}

func NewServicesConfig(indexerConfig *IndexerConfig) (*ServicesConfig, error) {
	externalServices, err := ParseExternalServices(indexerConfig)
	if err != nil {
		return nil, err
	}

	runServices, err := ParseRunServices(indexerConfig.RunServices)
	if err != nil {
		return nil, err
	}

	metrics, err := ParseMetrics(indexerConfig)
	if err != nil {
		return nil, err
	}

	servicesConfig := &ServicesConfig{
		ExternalServices: externalServices,
		RunServices:      runServices,
		Metrics:          metrics,
	}

	err = ValidateDependencies(servicesConfig)
	if err != nil {
		return nil, err
	}

	return servicesConfig, nil
}

func ParseExternalServices(indexerConfig *IndexerConfig) (*ExternalServicesConfig, error) {
	cfg := &ExternalServicesConfig{
		BitcoinRpcHost:            indexerConfig.BitcoinRpcHost,
		BitcoinRpcUser:            indexerConfig.BitcoinRpcUser,
		BitcoinRpcPass:            indexerConfig.BitcoinRpcPass,
		TonConfigUrl:              indexerConfig.TonConfigUrl,
		DatabaseUrl:               indexerConfig.DatabaseUrl,
		TeleportContractAddr:      nil,
		CoordinatorContractAddr:   nil,
		BitcoinClientContractAddr: nil,
		JettonMinterContractAddr:  nil,
	}

	var err error = nil
	if len(indexerConfig.TeleportContractAddr) > 0 {
		cfg.TeleportContractAddr, err = address.ParseAddr(indexerConfig.TeleportContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Teleport Contract address '%s' failed", indexerConfig.TeleportContractAddr)
		}
	}

	if len(indexerConfig.CoordinatorContractAddr) > 0 {
		cfg.CoordinatorContractAddr, err = address.ParseAddr(indexerConfig.CoordinatorContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Coordinator Contract address '%s' failed", indexerConfig.CoordinatorContractAddr)
		}
	}

	if len(indexerConfig.BitcoinClientContractAddr) > 0 {
		cfg.BitcoinClientContractAddr, err = address.ParseAddr(indexerConfig.BitcoinClientContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Bitcoin Client Contract address '%s' failed", indexerConfig.BitcoinClientContractAddr)
		}
	}

	if len(indexerConfig.JettonMinterContractAddr) > 0 {
		cfg.JettonMinterContractAddr, err = address.ParseAddr(indexerConfig.JettonMinterContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Jetton Minter Contract address '%s' failed", indexerConfig.JettonMinterContractAddr)
		}
	}

	return cfg, nil
}

func ParseRunServices(envStr string) (*RunServicesConfig, error) {
	cfg := &RunServicesConfig{
		RunMintService:    true,
		RunEventService:   true,
		RunPegoutManager:  true,
		RunHttpService:    true,
		RunMetricsService: true,
	}

	if len(envStr) == 0 {
		return cfg, nil
	}

	parts := strings.Split(envStr, ",")

	for _, part := range parts {
		subparts := strings.Split(part, "=")
		if len(subparts) != 2 {
			return nil, fmt.Errorf("wrong `RUN_SERVICES` .env argument '%s'", part)
		}

		name := subparts[0]
		value, err := ParseBool(subparts[1], name)
		if err != nil {
			return nil, fmt.Errorf("wrong `RUN_SERVICES` .env argument value '%s'. %w", part, err)
		}

		switch name {
		case "MINT_SERVICE":
			cfg.RunMintService = value
		case "EVENT_SERVICE":
			cfg.RunEventService = value
		case "PEGOUT_MANAGER":
			cfg.RunPegoutManager = value
		case "HTTP_SERVICE":
			cfg.RunHttpService = value
		case "METRICS_SERVICE":
			cfg.RunMetricsService = value
		default:
			return nil, fmt.Errorf("unknonwn `RUN_SERVICES` .env argument name '%s'", name)
		}
	}

	return cfg, nil
}

func ParseMetrics(indexerConfig *IndexerConfig) (*MetricsConfig, error) {
	cfg := &MetricsConfig{
		RunFetcherDKG:                   true,
		RunFetcherContractBalances:      true,
		RunFetcherContractBitcoinClient: true,
		RunFetcherContractTeleport:      true,
		RunFetcherContractCoordinator:   true,
		RunFetcherPegouts:               true,
		RunFetcherBitcoinNetwork:        true,

		WriterDbChainSize:                5,
		DkgFetchPeriod:                   10,
		BitcoinClientContractFetchPeriod: 60,
		BitcoinNetworkFetchPeriod:        59,
		TeleportContractFetchPeriod:      27,
		CoordinatorContractFetchPeriod:   59,
	}

	// Run fetchers
	if len(indexerConfig.Metrics) > 0 {
		envStr := indexerConfig.Metrics
		parts := strings.Split(envStr, ",")

		for _, part := range parts {
			subparts := strings.Split(part, "=")
			if len(subparts) != 2 {
				return nil, fmt.Errorf("wrong `METRICS` .env argument '%s'", part)
			}

			name := subparts[0]
			value, err := ParseBool(subparts[1], name)
			if err != nil {
				return nil, fmt.Errorf("wrong `METRICS` .env argument value '%s'. %w", part, err)
			}

			switch name {
			case "DKG":
				cfg.RunFetcherDKG = value
			case "CONTRACT_BALANCES":
				cfg.RunFetcherContractBalances = value
			case "CONTRACT_BITCOIN_CLIENT":
				cfg.RunFetcherContractBitcoinClient = value
			case "CONTRACT_TELEPORT":
				cfg.RunFetcherContractTeleport = value
			case "CONTRACT_COORDINATOR":
				cfg.RunFetcherContractCoordinator = value
			case "PEGOUTS":
				cfg.RunFetcherPegouts = value
			case "BITCOIN_NETWORK":
				cfg.RunFetcherBitcoinNetwork = value
			default:
				return nil, fmt.Errorf("unknonwn `METRICS` .env argument name '%s'", name)
			}
		}
	}

	// Fine tune arguments
	if len(indexerConfig.MetricsArgs) > 0 {
		envStr := indexerConfig.MetricsArgs
		parts := strings.Split(envStr, ",")

		for _, part := range parts {
			subparts := strings.Split(part, "=")
			if len(subparts) != 2 {
				return nil, fmt.Errorf("wrong `METRICS_ARGS` .env argument '%s'", part)
			}

			name := subparts[0]
			value, err := ParseInt(subparts[1], name)
			if err != nil {
				return nil, fmt.Errorf("wrong `METRICS_ARGS` .env argument value '%s'. %w", part, err)
			}

			switch name {
			case "WRITE_DB_CHAIN_SIZE":
				cfg.WriterDbChainSize = value
			case "DKG_FETCH_PERIOD":
				cfg.DkgFetchPeriod = value
			case "BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD":
				cfg.BitcoinClientContractFetchPeriod = value
			case "BITCOIN_NETWORK_FETCH_PERIOD":
				cfg.BitcoinNetworkFetchPeriod = value
			case "TELEPORT_CONTRACT_FETCH_PERIOD":
				cfg.TeleportContractFetchPeriod = value
			case "COORDINATOR_CONTRACT_FETCH_PERIOD":
				cfg.CoordinatorContractFetchPeriod = value

			default:
				return nil, fmt.Errorf("unknonwn `METRICS_ARGS` .env argument name '%s'", name)
			}
		}
	}

	return cfg, nil
}

func ParseBool(value string, name string) (bool, error) {
	if value == "true" {
		return true, nil
	}

	if value == "false" {
		return false, nil
	}

	return false, fmt.Errorf("incorrect boolean value '%s' assigned to %s", value, name)
}

func ParseInt(value string, name string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", value, name)
	}

	return int(val), nil
}

func ValidateDependencies(servicesConfig *ServicesConfig) error {
	return nil
}
