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
	HttpPort                         int
	AlertsTestApiEnable              bool
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
	AlertsCheckPeriod                int
}

func NewServicesConfig(config *EnvConfig) (*ServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8
	httpPort := 3000
	alertsTestApiEnable := false
	writerDbChainSize := 5
	dkgFetchPeriod := 10
	bitcoinClientContractFetchPeriod := 60
	bitcoinNetworkFetchPeriod := 59
	teleportContractFetchPeriod := 27
	coordinatorContractFetchPeriod := 12
	alertsCheckPeriod := 15

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

	if len(config.HttpPort) > 0 {
		value, err := ParseInt(config.HttpPort, "HttpPort")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_HTTP_PORT` .env argument value '%s'. %w", config.HttpPort, err)
		}

		httpPort = value
	}

	if len(config.AlertsTestApiEnable) > 0 {
		value, err := ParseBool(config.AlertsTestApiEnable, "AlertsTestApiEnable")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_ALERTS_TEST_API_ENABLE` .env argument value '%s'. %w", config.AlertsTestApiEnable, err)
		}

		alertsTestApiEnable = value
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

	if len(config.AlertsCheckPeriod) > 0 {
		value, err := ParseInt(config.AlertsCheckPeriod, "AlertsCheckPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `ALERTS_CHECK_PERIOD` .env argument value '%s'. %w", config.DatabaseMaxConn, err)
		}

		alertsCheckPeriod = value
	}

	servicesConfig := &ServicesConfig{
		BitcoinRpcHost:                   config.BitcoinRpcHost,
		BitcoinRpcUser:                   config.BitcoinRpcUser,
		BitcoinRpcPass:                   config.BitcoinRpcPass,
		TonConfigUrl:                     config.TonConfigUrl,
		DatabaseUrl:                      config.DatabaseUrl,
		DatabaseMaxConn:                  databaseMaxConn,
		DatabaseMaxIdleConn:              databaseMaxIdleConn,
		HttpPort:                         httpPort,
		AlertsTestApiEnable:              alertsTestApiEnable,
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
		AlertsCheckPeriod:                alertsCheckPeriod,
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
HttpPort: %d                        ,
AlertsTestApiEnable: %t,
WriterDbChainSize: %d
DkgFetchPeriod: %d sec.
BitcoinClientContractFetchPeriod: %d sec.
BitcoinNetworkFetchPeriod: %d sec.
TeleportContractFetchPeriod: %d sec.
CoordinatorContractFetchPeriod: %d sec.
AlertsCheckPeriod: %d sec.
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.TeleportContractAddr,
		config.CoordinatorContractAddr,
		config.BitcoinClientContractAddr,
		config.JettonMinterContractAddr,
		config.DatabaseMaxConn,
		config.DatabaseMaxIdleConn,
		config.HttpPort,
		config.AlertsTestApiEnable,
		config.WriterDbChainSize,
		config.DkgFetchPeriod,
		config.BitcoinClientContractFetchPeriod,
		config.BitcoinNetworkFetchPeriod,
		config.TeleportContractFetchPeriod,
		config.CoordinatorContractFetchPeriod,
		config.AlertsCheckPeriod,
	)
}

func ParseInt(value string, name string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("incorrect int value '%s' assigned to %s", value, name)
	}

	return int(val), nil
}

func ParseBool(value string, name string) (bool, error) {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("incorrect bool value '%s' assigned to %s", value, name)
	}

	return val, nil
}
