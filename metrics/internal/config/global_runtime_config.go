package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type GlobalRuntimeConfig struct {
	tonClient *tonclient.TonClient

	mu                   sync.RWMutex
	tonMaxMainValidators int
}

func NewGlobalRuntimeConfig(
	tonClient *tonclient.TonClient,
) *GlobalRuntimeConfig {
	return &GlobalRuntimeConfig{
		tonClient:            tonClient,
		tonMaxMainValidators: -1,
	}
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
