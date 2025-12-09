package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type GlobalRuntimeConfig struct {
	tonClient   *tonclient.TonClient
	tonExplorer string
	btcExplorer string
	runbook     string

	mu                   sync.RWMutex
	tonMaxMainValidators int
}

var globalRuntimeConfigG *GlobalRuntimeConfig = nil

func InitGlobalRuntimeConfig(
	tonClient *tonclient.TonClient,
	cfg *ServicesConfig,
) {
	globalRuntimeConfigG = &GlobalRuntimeConfig{
		tonClient:            tonClient,
		tonMaxMainValidators: -1,
		tonExplorer:          cfg.TonExplorer,
		btcExplorer:          cfg.BtcExplorer,
		runbook:              cfg.Runbook,
	}
}

func GetGlobalRuntimeConfig() *GlobalRuntimeConfig {
	return globalRuntimeConfigG
}

func (cfg *GlobalRuntimeConfig) TonMaxMainValidators(ctx context.Context) (int, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.tonMaxMainValidators < 0 {
		block, err := cfg.tonClient.API.GetMasterchainInfo(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get block: %v", err)
		}

		tonConfig, err := cfg.tonClient.API.GetBlockchainConfig(ctx, block, 16)
		if err != nil {
			return 0, fmt.Errorf("failed to get config: %v", err)
		}

		tonConfigParam16 := tonConfig.Get(16)
		s := tonConfigParam16.BeginParse()
		s.MustLoadUInt(16)
		cfg.tonMaxMainValidators = int(s.MustLoadUInt(16))
	}

	return cfg.tonMaxMainValidators, nil
}

func (cfg *GlobalRuntimeConfig) TonExplorer() string {
	return cfg.tonExplorer
}

func (cfg *GlobalRuntimeConfig) BtcExplorer() string {
	return cfg.btcExplorer
}

func (cfg *GlobalRuntimeConfig) RunbookUrl() string {
	return cfg.runbook
}
