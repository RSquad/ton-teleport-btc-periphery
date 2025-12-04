package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
)

type FetcherContractCoordinator struct {
	chDB                chan MetricsPayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
}

func NewFetcherContractCoordinator(
	chDB chan MetricsPayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
) *FetcherContractCoordinator {
	return &FetcherContractCoordinator{
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
	}
}

func (fetcher *FetcherContractCoordinator) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("FetcherContractCoordinator: stopped")
	logger.DefaultLogStartWork("FetcherContractCoordinator: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch("FetcherContractCoordinator", time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch("FetcherContractCoordinator")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherContractCoordinator received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
			watchdog.Global().Heartbeat("FetcherContractCoordinator")
		}
	}
}

func (fetcher *FetcherContractCoordinator) Fetch() {
	storage, err := fetcher.coordinatorContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractCoordinator: failed to retrieve storage cell, error: %v", err))
		return
	}

	storage.Dkg = nil
	storage.PrevDkg = nil

	jsonData, err := json.Marshal(storage)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractCoordinator").
			Msg("failed to serialize FetcherContractCoordinator->json")

		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeContractCoordinator,
		payload: string(jsonData),
	}
}
