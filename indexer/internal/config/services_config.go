package config

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

type ServicesConfig struct {
	BitcoinRpcHost            string
	BitcoinRpcUser            string
	BitcoinRpcPass            string
	TonConfigUrl              string
	DatabaseUrl               string
	DatabaseMaxConn           int
	DatabaseMaxIdleConn       int
	CoordinatorContractAddr   *address.Address
	PProfHttpEnable           bool
	ServerTimeout             time.Duration
	BitcoinClientContractAddr *address.Address
	IndexerWalletV4Secret     string
	HighLoadWalletV3Seed      string
}

var (
	globalConfig     *ServicesConfig
	globalConfigOnce sync.Once
)

func initGlobalConfig(cfg *ServicesConfig) {
	globalConfigOnce.Do(func() {
		globalConfig = cfg
	})
}

func Get() *ServicesConfig {
	return globalConfig
}

func NewServicesConfig(envConfig *EnvConfig) (*ServicesConfig, error) {
	databaseMaxConn := 8
	databaseMaxIdleConn := 8
	pprofHttpEnable := false

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

	BitcoinClientContractAddr, err := address.ParseAddr(envConfig.BitcoinClientContractAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Bitcoin Client Contract address '%s' failed", envConfig.CoordinatorContractAddr)
	}

	if len(envConfig.PProfHttpEnable) > 0 {
		value, err := ParseBool(envConfig.PProfHttpEnable, "PProfHttpEnable")
		if err != nil {
			return nil, fmt.Errorf("wrong `INDEXER_PPROF_HTTP_ENABLE` .env argument value '%s'. %w", envConfig.PProfHttpEnable, err)
		}

		pprofHttpEnable = value
	}

	cfg := &ServicesConfig{
		BitcoinRpcHost:            envConfig.BitcoinRpcHost,
		BitcoinRpcUser:            envConfig.BitcoinRpcUser,
		BitcoinRpcPass:            envConfig.BitcoinRpcPass,
		TonConfigUrl:              envConfig.TonConfigUrl,
		DatabaseUrl:               envConfig.DatabaseUrl,
		DatabaseMaxConn:           databaseMaxConn,
		DatabaseMaxIdleConn:       databaseMaxIdleConn,
		CoordinatorContractAddr:   coordinatorContractAddr,
		PProfHttpEnable:           pprofHttpEnable,
		ServerTimeout:             envConfig.ServerTimeout,
		BitcoinClientContractAddr: BitcoinClientContractAddr,
		IndexerWalletV4Secret:     envConfig.IndexerWalletV4Secret,
		HighLoadWalletV3Seed:      envConfig.HighLoadWalletV3Secret,
	}

	initGlobalConfig(cfg)
	return cfg, nil
}

func CfgToString(config *ServicesConfig) string {
	return fmt.Sprintf(
		`BitcoinRpcHost: %s
TonConfigUrl: %s
DatabaseMaxConn: %d
DatabaseMaxIdleConn: %d
CoordinatorContractAddr: %s
PProfHttpEnable: %t
`,
		config.BitcoinRpcHost,
		config.TonConfigUrl,
		config.DatabaseMaxConn,
		config.DatabaseMaxIdleConn,
		config.CoordinatorContractAddr,
		config.PProfHttpEnable,
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
