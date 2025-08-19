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
	MetricsPegoutsFetchPeriod        int
}

type ExternalServicesConfig struct {
	BitcoinRpcHost            string
	BitcoinRpcUser            string
	BitcoinRpcPass            string
	TonConfigUrl              string
	DatabaseUrl               string
	DatabaseMaxConn           int
	DatabaseMaxIdleConn       int
	RelayerWalletV4Secret     string
	TeleportContractAddr      *address.Address
	CoordinatorContractAddr   *address.Address
	BitcoinClientContractAddr *address.Address
	JettonMinterContractAddr  *address.Address
}

type ServicesConfig struct {
	ExternalServices *ExternalServicesConfig
	Metrics          *MetricsConfig
}

func NewServicesConfig(config *Config) (*ServicesConfig, error) {
	externalServices, err := ParseExternalServices(config)
	if err != nil {
		return nil, err
	}

	metrics, err := ParseMetrics(config)
	if err != nil {
		return nil, err
	}

	servicesConfig := &ServicesConfig{
		ExternalServices: externalServices,
		Metrics:          metrics,
	}

	err = ValidateDependencies(servicesConfig)
	if err != nil {
		return nil, err
	}

	return servicesConfig, nil
}

func CfgToString(config *Config) string {
	return fmt.Sprintf(
		`BitcoinRpcHost: %s
TonConfigUrl: %s
TeleportContractAddr: %s
CoordinatorContractAddr: %s
BitcoinClientContractAddr: %s
JettonMinterContractAddr: %s
Metrics: %s
MetricsArgs: %s
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.TeleportContractAddr,
		config.CoordinatorContractAddr,
		config.BitcoinClientContractAddr,
		config.JettonMinterContractAddr,
		config.Metrics,
		config.MetricsArgs,
	)
}

func ParseExternalServices(config *Config) (*ExternalServicesConfig, error) {
	cfg := &ExternalServicesConfig{
		BitcoinRpcHost:            config.BitcoinRpcHost,
		BitcoinRpcUser:            config.BitcoinRpcUser,
		BitcoinRpcPass:            config.BitcoinRpcPass,
		TonConfigUrl:              config.TonConfigUrl,
		DatabaseUrl:               config.DatabaseUrl,
		DatabaseMaxConn:           8,
		DatabaseMaxIdleConn:       8,
		TeleportContractAddr:      nil,
		CoordinatorContractAddr:   nil,
		BitcoinClientContractAddr: nil,
		JettonMinterContractAddr:  nil,
	}

	if len(config.DatabaseMaxConn) > 0 {
		value, err := ParseInt(config.DatabaseMaxConn, "DatabaseMaxConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `DATABASE_MAX_CONN` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		cfg.DatabaseMaxConn = value
	}

	if len(config.DatabaseMaxIdleConn) > 0 {
		value, err := ParseInt(config.DatabaseMaxIdleConn, "DatabaseMaxIdleConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `DATABASE_MAX_IDLE_CONN` .env argument value '%s'. %w", config.DatabaseMaxIdleConn, err)
		}

		cfg.DatabaseMaxIdleConn = value
	}

	var err error = nil
	if len(config.TeleportContractAddr) > 0 {
		cfg.TeleportContractAddr, err = address.ParseAddr(config.TeleportContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Teleport Contract address '%s' failed", config.TeleportContractAddr)
		}
	}

	if len(config.CoordinatorContractAddr) > 0 {
		cfg.CoordinatorContractAddr, err = address.ParseAddr(config.CoordinatorContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Coordinator Contract address '%s' failed", config.CoordinatorContractAddr)
		}
	}

	if len(config.BitcoinClientContractAddr) > 0 {
		cfg.BitcoinClientContractAddr, err = address.ParseAddr(config.BitcoinClientContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Bitcoin Client Contract address '%s' failed", config.BitcoinClientContractAddr)
		}
	}

	if len(config.JettonMinterContractAddr) > 0 {
		cfg.JettonMinterContractAddr, err = address.ParseAddr(config.JettonMinterContractAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Jetton Minter Contract address '%s' failed", config.JettonMinterContractAddr)
		}
	}

	return cfg, nil
}

func ParseMetrics(config *Config) (*MetricsConfig, error) {
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
		MetricsPegoutsFetchPeriod:        59,
	}

	// Run fetchers
	if len(config.Metrics) > 0 {
		envStr := config.Metrics
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
	if len(config.MetricsArgs) > 0 {
		envStr := config.MetricsArgs
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
			case "METRICS_PEGOUTS_FETCH_PERIOD":
				cfg.MetricsPegoutsFetchPeriod = value

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
