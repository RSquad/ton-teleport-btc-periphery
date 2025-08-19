package config

import (
	"fmt"
	"strconv"

	"github.com/xssnick/tonutils-go/address"
)

type ServicesConfig struct {
	BitcoinRpcHost          string
	BitcoinRpcUser          string
	BitcoinRpcPass          string
	TonConfigUrl            string
	DatabaseUrl             string
	DatabaseMaxConn         int
	DatabaseMaxIdleConn     int
	CoordinatorContractAddr *address.Address
}

func NewServicesConfig(envConfig *EnvConfig) (*ServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8

	if len(envConfig.DatabaseMaxConn) > 0 {
		value, err := ParseInt(envConfig.DatabaseMaxConn, "DatabaseMaxConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `INDEXER_DATABASE_MAX_CONN` .env argument value '%s'. %w", envConfig.DatabaseMaxConn, err)
		}
		databaseMaxConn = value
	}

	if len(envConfig.DatabaseMaxIdleConn) > 0 {
		value, err := ParseInt(envConfig.DatabaseMaxIdleConn, "DatabaseMaxIdleConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `INDEXER_DATABASE_MAX_IDLE_CONN` .env argument value '%s'. %w", envConfig.DatabaseMaxIdleConn, err)
		}
		databaseMaxIdleConn = value
	}

	coordinatorContractAddr, err := address.ParseAddr(envConfig.CoordinatorContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Coordinator Contract address '%s' failed", envConfig.CoordinatorContractAddr)
	}

	cfg := &ServicesConfig{
		BitcoinRpcHost:          envConfig.BitcoinRpcHost,
		BitcoinRpcUser:          envConfig.BitcoinRpcUser,
		BitcoinRpcPass:          envConfig.BitcoinRpcPass,
		TonConfigUrl:            envConfig.TonConfigUrl,
		DatabaseUrl:             envConfig.DatabaseUrl,
		DatabaseMaxConn:         databaseMaxConn,
		DatabaseMaxIdleConn:     databaseMaxIdleConn,
		CoordinatorContractAddr: coordinatorContractAddr,
	}

	return cfg, nil
}

func CfgToString(config *ServicesConfig) string {
	return fmt.Sprintf(
		`BitcoinRpcHost: %s
TonConfigUrl: %s
DatabaseMaxConn: %d
DatabaseMaxIdleConn: %d
CoordinatorContractAddr: %s
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.DatabaseMaxConn,
		config.DatabaseMaxIdleConn,
		config.CoordinatorContractAddr,
	)
}

func ParseInt(value string, name string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", value, name)
	}

	return int(val), nil
}
