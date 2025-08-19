package config

import (
	"fmt"
	"strconv"

	"github.com/xssnick/tonutils-go/address"
)

type ExternalServicesConfig struct {
	BitcoinRpcHost            string
	BitcoinRpcUser            string
	BitcoinRpcPass            string
	TonConfigUrl              string
	DatabaseUrl               string
	DatabaseMaxConn           int
	DatabaseMaxIdleConn       int
	TeleportContractAddr      *address.Address
	CoordinatorContractAddr   *address.Address
	BitcoinClientContractAddr *address.Address
	JettonMinterContractAddr  *address.Address
}

type ServicesConfig struct {
	ExternalServices *ExternalServicesConfig
}

func NewServicesConfig(indexerConfig *IndexerConfig) (*ServicesConfig, error) {
	externalServices, err := ParseExternalServices(indexerConfig)
	if err != nil {
		return nil, err
	}

	servicesConfig := &ServicesConfig{
		ExternalServices: externalServices,
	}

	return servicesConfig, nil
}

func CfgToString(indexerConfig *IndexerConfig) string {
	return fmt.Sprintf(
		`BitcoinRpcHost: %s
TonConfigUrl: %s
TeleportContractAddr: %s
CoordinatorContractAddr: %s
BitcoinClientContractAddr: %s
JettonMinterContractAddr: %s
`,
		indexerConfig.BitcoinRpcHost,
		indexerConfig.TonConfigUrl,
		indexerConfig.TeleportContractAddr,
		indexerConfig.CoordinatorContractAddr,
		indexerConfig.BitcoinClientContractAddr,
		indexerConfig.JettonMinterContractAddr,
	)
}

func ParseExternalServices(indexerConfig *IndexerConfig) (*ExternalServicesConfig, error) {
	cfg := &ExternalServicesConfig{
		BitcoinRpcHost:            indexerConfig.BitcoinRpcHost,
		BitcoinRpcUser:            indexerConfig.BitcoinRpcUser,
		BitcoinRpcPass:            indexerConfig.BitcoinRpcPass,
		TonConfigUrl:              indexerConfig.TonConfigUrl,
		DatabaseUrl:               indexerConfig.DatabaseUrl,
		DatabaseMaxConn:           8,
		DatabaseMaxIdleConn:       8,
		TeleportContractAddr:      nil,
		CoordinatorContractAddr:   nil,
		BitcoinClientContractAddr: nil,
		JettonMinterContractAddr:  nil,
	}

	if len(indexerConfig.DatabaseMaxConn) > 0 {
		value, err := ParseInt(indexerConfig.DatabaseMaxConn, "DatabaseMaxConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `INDEXER_DATABASE_MAX_CONN` .env argument value '%s'. %w", indexerConfig.DatabaseMaxConn, err)
		}

		cfg.DatabaseMaxConn = value
	}

	if len(indexerConfig.DatabaseMaxIdleConn) > 0 {
		value, err := ParseInt(indexerConfig.DatabaseMaxIdleConn, "DatabaseMaxIdleConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `INDEXER_DATABASE_MAX_IDLE_CONN` .env argument value '%s'. %w", indexerConfig.DatabaseMaxIdleConn, err)
		}

		cfg.DatabaseMaxIdleConn = value
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

func ParseInt(value string, name string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", value, name)
	}

	return int(val), nil
}
