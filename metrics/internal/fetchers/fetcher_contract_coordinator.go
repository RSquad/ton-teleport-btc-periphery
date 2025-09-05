package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type FetcherContractCoordinator struct {
	chDB                chan PayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
}

func NewFetcherContractCoordinator(
	chDB chan PayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
) *FetcherContractCoordinator {
	return &FetcherContractCoordinator{
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
	}
}

func (fetcher *FetcherContractCoordinator) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractCoordinator: stopped")
	logger.DefaultLogStartWork("FetcherContractCoordinator: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherContractCoordinator received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
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
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeContractCoordinator,
		payload: string(jsonData),
	}
}
