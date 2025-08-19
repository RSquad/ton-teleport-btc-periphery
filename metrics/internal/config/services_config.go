package config

import (
	"fmt"
	"strconv"

	"github.com/xssnick/tonutils-go/address"
)

type ServicesConfig struct {
	BitcoinRpcHost                   string
	BitcoinRpcUser                   string
	BitcoinRpcPass                   string
	TonConfigUrl                     string
	DatabaseUrl                      string
	DatabaseMaxConn                  int
	DatabaseMaxIdleConn              int
	TeleportContractAddr             *address.Address
	CoordinatorContractAddr          *address.Address
	BitcoinClientContractAddr        *address.Address
	JettonMinterContractAddr         *address.Address
	WriterDbChainSize                int
	DkgFetchPeriod                   int
	BitcoinClientContractFetchPeriod int
	BitcoinNetworkFetchPeriod        int
	TeleportContractFetchPeriod      int
	CoordinatorContractFetchPeriod   int
	MetricsPegoutsFetchPeriod        int
}

func NewServicesConfig(config *EnvConfig) (*ServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8
	writerDbChainSize := 5
	dkgFetchPeriod := 10
	bitcoinClientContractFetchPeriod := 60
	bitcoinNetworkFetchPeriod := 59
	teleportContractFetchPeriod := 27
	coordinatorContractFetchPeriod := 12
	metricsPegoutsFetchPeriod := 59

	if len(config.DatabaseMaxConn) > 0 {
		value, err := ParseInt(config.DatabaseMaxConn, "DatabaseMaxConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_DATABASE_MAX_CONN` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		databaseMaxConn = value
	}

	if len(config.DatabaseMaxIdleConn) > 0 {
		value, err := ParseInt(config.DatabaseMaxIdleConn, "DatabaseMaxIdleConn")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_DATABASE_MAX_IDLE_CONN` .env argument value '%s'. %w", config.DatabaseMaxIdleConn, err)
		}

		databaseMaxIdleConn = value
	}

	teleportContractAddr, err := address.ParseAddr(config.TeleportContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Teleport Contract address '%s' failed", config.TeleportContractAddr)
	}

	coordinatorContractAddr, err := address.ParseAddr(config.CoordinatorContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Coordinator Contract address '%s' failed", config.CoordinatorContractAddr)
	}

	bitcoinClientContractAddr, err := address.ParseAddr(config.BitcoinClientContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Bitcoin Client Contract address '%s' failed", config.BitcoinClientContractAddr)
	}

	jettonMinterContractAddr, err := address.ParseAddr(config.JettonMinterContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Jetton Minter Contract address '%s' failed", config.JettonMinterContractAddr)
	}

	if len(config.WriterDbChainSize) > 0 {
		value, err := ParseInt(config.WriterDbChainSize, "WriterDbChainSize")
		if err != nil {
			return nil, fmt.Errorf("wrong `WRITE_DB_CHAIN_SIZE` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		writerDbChainSize = value
	}

	if len(config.DkgFetchPeriod) > 0 {
		value, err := ParseInt(config.DkgFetchPeriod, "DkgFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `DKG_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		dkgFetchPeriod = value
	}

	if len(config.BitcoinClientContractFetchPeriod) > 0 {
		value, err := ParseInt(config.BitcoinClientContractFetchPeriod, "BitcoinClientContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		bitcoinClientContractFetchPeriod = value
	}

	if len(config.BitcoinNetworkFetchPeriod) > 0 {
		value, err := ParseInt(config.BitcoinNetworkFetchPeriod, "BitcoinNetworkFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `BITCOIN_NETWORK_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		bitcoinNetworkFetchPeriod = value
	}

	if len(config.TeleportContractFetchPeriod) > 0 {
		value, err := ParseInt(config.TeleportContractFetchPeriod, "TeleportContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `TELEPORT_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		teleportContractFetchPeriod = value
	}

	if len(config.CoordinatorContractFetchPeriod) > 0 {
		value, err := ParseInt(config.CoordinatorContractFetchPeriod, "CoordinatorContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `COORDINATOR_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		coordinatorContractFetchPeriod = value
	}

	if len(config.MetricsPegoutsFetchPeriod) > 0 {
		value, err := ParseInt(config.MetricsPegoutsFetchPeriod, "MetricsPegoutsFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_PEGOUTS_FETCH_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		metricsPegoutsFetchPeriod = value
	}

	servicesConfig := &ServicesConfig{
		BitcoinRpcHost:                   config.BitcoinRpcHost,
		BitcoinRpcUser:                   config.BitcoinRpcUser,
		BitcoinRpcPass:                   config.BitcoinRpcPass,
		TonConfigUrl:                     config.TonConfigUrl,
		DatabaseUrl:                      config.DatabaseUrl,
		DatabaseMaxConn:                  databaseMaxConn,
		DatabaseMaxIdleConn:              databaseMaxIdleConn,
		TeleportContractAddr:             teleportContractAddr,
		CoordinatorContractAddr:          coordinatorContractAddr,
		BitcoinClientContractAddr:        bitcoinClientContractAddr,
		JettonMinterContractAddr:         jettonMinterContractAddr,
		WriterDbChainSize:                writerDbChainSize,
		DkgFetchPeriod:                   dkgFetchPeriod,
		BitcoinClientContractFetchPeriod: bitcoinClientContractFetchPeriod,
		BitcoinNetworkFetchPeriod:        bitcoinNetworkFetchPeriod,
		TeleportContractFetchPeriod:      teleportContractFetchPeriod,
		CoordinatorContractFetchPeriod:   coordinatorContractFetchPeriod,
		MetricsPegoutsFetchPeriod:        metricsPegoutsFetchPeriod,
	}

	return servicesConfig, nil
}

func CfgToString(config *ServicesConfig) string {
	return fmt.Sprintf(
		`BitcoinRpcHost: %s
TonConfigUrl: %s
TeleportContractAddr: %s
CoordinatorContractAddr: %s
BitcoinClientContractAddr: %s
JettonMinterContractAddr: %s
DatabaseMaxConn: %d
DatabaseMaxIdleConn: %d
WriterDbChainSize: %d
DkgFetchPeriod: %d
BitcoinClientContractFetchPeriod: %d
BitcoinNetworkFetchPeriod: %d
TeleportContractFetchPeriod: %d
CoordinatorContractFetchPeriod: %d
MetricsPegoutsFetchPeriod: %d
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.TeleportContractAddr,
		config.CoordinatorContractAddr,
		config.BitcoinClientContractAddr,
		config.JettonMinterContractAddr,
		config.DatabaseMaxConn,
		config.DatabaseMaxIdleConn,
		config.WriterDbChainSize,
		config.DkgFetchPeriod,
		config.BitcoinClientContractFetchPeriod,
		config.BitcoinNetworkFetchPeriod,
		config.TeleportContractFetchPeriod,
		config.CoordinatorContractFetchPeriod,
		config.MetricsPegoutsFetchPeriod,
	)
}

func ParseInt(value string, name string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", value, name)
	}

	return int(val), nil
}
