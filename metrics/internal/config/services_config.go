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
	TeleportContractAddr             *address.Address
	CoordinatorContractAddr          *address.Address
	BitcoinClientContractAddr        *address.Address
	JettonMinterContractAddr         *address.Address
	RelayerWalletAddr                *address.Address
	WriterDbChainSize                int
	DkgFetchPeriod                   int
	BitcoinClientContractFetchPeriod int
	BitcoinNetworkFetchPeriod        int
	TeleportContractFetchPeriod      int
	CoordinatorContractFetchPeriod   int
	ContractBalancesFetchPeriod      int
	PProfHttpEnable                  bool
	TonExplorer                      string
	BtcExplorer                      string
	Runbook                          string
	AlertsTestApiEnable              bool
	AlertsCheckPeriod                int
	AlertBtcBlockDeltaHeightWarn     int
	AlertBtcBlockDeltaHeightCrit     int
}

func NewServicesConfig(config *EnvConfig) (*ServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8
	httpPort := 3000
	writerDbChainSize := 5
	dkgFetchPeriod := 10
	bitcoinClientContractFetchPeriod := 60
	bitcoinNetworkFetchPeriod := 59
	teleportContractFetchPeriod := 27
	coordinatorContractFetchPeriod := 12
	contractBalancesFetchPeriod := 150
	pprofHttpEnable := false
	alertsTestApiEnable := false
	alertsCheckPeriod := 15
	alertBtcBlockDeltaHeightWarn := 3
	alertBtcBlockDeltaHeightCrit := 4

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

	relayerWalletAddr, err := address.ParseAddr(config.RelayerWalletAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Relayer Wallet address '%s' failed", config.RelayerWalletAddr)
	}

	if len(config.WriterDbChainSize) > 0 {
		value, err := ParseInt(config.WriterDbChainSize, "WriterDbChainSize")
		if err != nil {
			return nil, fmt.Errorf("wrong `WRITE_DB_CHAIN_SIZE` .env argument value '%s'. %w", config.WriterDbChainSize, err)
		}

		writerDbChainSize = value
	}

	if len(config.DkgFetchPeriod) > 0 {
		value, err := ParseInt(config.DkgFetchPeriod, "DkgFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `DKG_FETCH_PERIOD` .env argument value '%s'. %w", config.DkgFetchPeriod, err)
		}

		dkgFetchPeriod = value
	}

	if len(config.BitcoinClientContractFetchPeriod) > 0 {
		value, err := ParseInt(config.BitcoinClientContractFetchPeriod, "BitcoinClientContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `BITCOIN_CLIENT_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.BitcoinClientContractFetchPeriod, err)
		}

		bitcoinClientContractFetchPeriod = value
	}

	if len(config.BitcoinNetworkFetchPeriod) > 0 {
		value, err := ParseInt(config.BitcoinNetworkFetchPeriod, "BitcoinNetworkFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `BITCOIN_NETWORK_FETCH_PERIOD` .env argument value '%s'. %w", config.BitcoinNetworkFetchPeriod, err)
		}

		bitcoinNetworkFetchPeriod = value
	}

	if len(config.TeleportContractFetchPeriod) > 0 {
		value, err := ParseInt(config.TeleportContractFetchPeriod, "TeleportContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `TELEPORT_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.TeleportContractFetchPeriod, err)
		}

		teleportContractFetchPeriod = value
	}

	if len(config.CoordinatorContractFetchPeriod) > 0 {
		value, err := ParseInt(config.CoordinatorContractFetchPeriod, "CoordinatorContractFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `COORDINATOR_CONTRACT_FETCH_PERIOD` .env argument value '%s'. %w", config.CoordinatorContractFetchPeriod, err)
		}

		coordinatorContractFetchPeriod = value
	}

	if len(config.ContractBalancesFetchPeriod) > 0 {
		value, err := ParseInt(config.ContractBalancesFetchPeriod, "ContractBalancesFetchPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `CONTRACT_BALANCES_FETCH_PERIOD` .env argument value '%s'. %w", config.ContractBalancesFetchPeriod, err)
		}

		contractBalancesFetchPeriod = value
	}

	if len(config.PProfHttpEnable) > 0 {
		value, err := ParseBool(config.PProfHttpEnable, "PProfHttpEnable")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_PPROF_HTTP_ENABLE` .env argument value '%s'. %w", config.PProfHttpEnable, err)
		}

		pprofHttpEnable = value
	}

	if len(config.AlertsTestApiEnable) > 0 {
		value, err := ParseBool(config.AlertsTestApiEnable, "AlertsTestApiEnable")
		if err != nil {
			return nil, fmt.Errorf("wrong `METRICS_ALERTS_TEST_API_ENABLE` .env argument value '%s'. %w", config.AlertsTestApiEnable, err)
		}

		alertsTestApiEnable = value
	}

	if len(config.AlertsCheckPeriod) > 0 {
		value, err := ParseInt(config.AlertsCheckPeriod, "AlertsCheckPeriod")
		if err != nil {
			return nil, fmt.Errorf("wrong `ALERTS_CHECK_PERIOD` .env argument value '%s'. %w", config.AlertsCheckPeriod, err)
		}

		alertsCheckPeriod = value
	}

	if len(config.AlertBtcBlockDeltaHeightWarn) > 0 {
		value, err := ParseInt(config.AlertBtcBlockDeltaHeightWarn, "AlertBtcBlockDeltaHeightWarn")
		if err != nil {
			return nil, fmt.Errorf("wrong `ALERT_BTC_BLOCK_DELTA_HEIGHT_WARN` .env argument value '%s'. %w", config.AlertBtcBlockDeltaHeightWarn, err)
		}

		alertBtcBlockDeltaHeightWarn = value
	}

	if len(config.AlertBtcBlockDeltaHeightCrit) > 0 {
		value, err := ParseInt(config.AlertBtcBlockDeltaHeightCrit, "AlertBtcBlockDeltaHeightCrit")
		if err != nil {
			return nil, fmt.Errorf("wrong `ALERT_BTC_BLOCK_DELTA_HEIGHT_CRIT` .env argument value '%s'. %w", config.AlertBtcBlockDeltaHeightCrit, err)
		}

		alertBtcBlockDeltaHeightCrit = value
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
		TeleportContractAddr:             teleportContractAddr,
		CoordinatorContractAddr:          coordinatorContractAddr,
		BitcoinClientContractAddr:        bitcoinClientContractAddr,
		JettonMinterContractAddr:         jettonMinterContractAddr,
		RelayerWalletAddr:                relayerWalletAddr,
		WriterDbChainSize:                writerDbChainSize,
		DkgFetchPeriod:                   dkgFetchPeriod,
		BitcoinClientContractFetchPeriod: bitcoinClientContractFetchPeriod,
		BitcoinNetworkFetchPeriod:        bitcoinNetworkFetchPeriod,
		TeleportContractFetchPeriod:      teleportContractFetchPeriod,
		CoordinatorContractFetchPeriod:   coordinatorContractFetchPeriod,
		ContractBalancesFetchPeriod:      contractBalancesFetchPeriod,
		PProfHttpEnable:                  pprofHttpEnable,
		TonExplorer:                      config.TonExplorer,
		BtcExplorer:                      config.BtcExplorer,
		Runbook:                          config.Runbook,
		AlertsTestApiEnable:              alertsTestApiEnable,
		AlertsCheckPeriod:                alertsCheckPeriod,
		AlertBtcBlockDeltaHeightWarn:     alertBtcBlockDeltaHeightWarn,
		AlertBtcBlockDeltaHeightCrit:     alertBtcBlockDeltaHeightCrit,
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
RelayerWalletAddr: %s
DatabaseMaxConn: %d
DatabaseMaxIdleConn: %d
HttpPort: %d
WriterDbChainSize: %d
DkgFetchPeriod: %d sec.
BitcoinClientContractFetchPeriod: %d sec.
BitcoinNetworkFetchPeriod: %d sec.
TeleportContractFetchPeriod: %d sec.
CoordinatorContractFetchPeriod: %d sec.
ContractBalancesFetchPeriod: %d sec.
PProfHttpEnable: %t
AlertsTestApiEnable: %t
AlertsCheckPeriod: %d sec.
AlertBtcBlockDeltaHeightWarn: %d
AlertBtcBlockDeltaHeightCrit: %d
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.TeleportContractAddr,
		config.CoordinatorContractAddr,
		config.BitcoinClientContractAddr,
		config.JettonMinterContractAddr,
		config.RelayerWalletAddr,
		config.DatabaseMaxConn,
		config.DatabaseMaxIdleConn,
		config.HttpPort,
		config.WriterDbChainSize,
		config.DkgFetchPeriod,
		config.BitcoinClientContractFetchPeriod,
		config.BitcoinNetworkFetchPeriod,
		config.TeleportContractFetchPeriod,
		config.CoordinatorContractFetchPeriod,
		config.ContractBalancesFetchPeriod,
		config.PProfHttpEnable,
		config.AlertsTestApiEnable,
		config.AlertsCheckPeriod,
		config.AlertBtcBlockDeltaHeightWarn,
		config.AlertBtcBlockDeltaHeightCrit,
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
