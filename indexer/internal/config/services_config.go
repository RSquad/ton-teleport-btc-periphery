package config

import (
	"fmt"
	"strconv"

	"github.com/xssnick/tonutils-go/address"
)

type ExternalServicesConfig struct {
	BitcoinRpcHost          string
	BitcoinRpcUser          string
	BitcoinRpcPass          string
	TonConfigUrl            string
	DatabaseUrl             string
	DatabaseMaxConn         int
	DatabaseMaxIdleConn     int
	CoordinatorContractAddr *address.Address
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
CoordinatorContractAddr: %s
`,
		indexerConfig.BitcoinRpcHost,
		indexerConfig.TonConfigUrl,
		indexerConfig.CoordinatorContractAddr,
	)
}

func ParseExternalServices(indexerConfig *IndexerConfig) (*ExternalServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8

	if len(indexerConfig.DatabaseMaxConn) > 0 {
		value, err := ParseInt(indexerConfig.DatabaseMaxConn, "DatabaseMaxConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `DATABASE_MAX_CONN` .env argument value '%s'. %w", indexerConfig.DatabaseMaxConn, err)
		}
		databaseMaxConn = value
	}

	if len(indexerConfig.DatabaseMaxIdleConn) > 0 {
		value, err := ParseInt(indexerConfig.DatabaseMaxIdleConn, "DatabaseMaxIdleConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `DATABASE_MAX_IDLE_CONN` .env argument value '%s'. %w", indexerConfig.DatabaseMaxIdleConn, err)
		}
		databaseMaxIdleConn = value
	}

	coordinatorContractAddr, err := address.ParseAddr(indexerConfig.CoordinatorContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Coordinator Contract address '%s' failed", indexerConfig.CoordinatorContractAddr)
	}

	cfg := &ExternalServicesConfig{
		BitcoinRpcHost:          indexerConfig.BitcoinRpcHost,
		BitcoinRpcUser:          indexerConfig.BitcoinRpcUser,
		BitcoinRpcPass:          indexerConfig.BitcoinRpcPass,
		TonConfigUrl:            indexerConfig.TonConfigUrl,
		DatabaseUrl:             indexerConfig.DatabaseUrl,
		DatabaseMaxConn:         databaseMaxConn,
		DatabaseMaxIdleConn:     databaseMaxIdleConn,
		CoordinatorContractAddr: coordinatorContractAddr,
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
