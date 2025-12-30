package fetchers

import (
	"context"
	"encoding/json"
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
	component := "FetcherContractCoordinator"

	logger.Log.Info().
		Str("component", component).
		Msg("started")

	defer func() {
		logger.Log.Info().
			Str("component", component).
			Msg("finished")
	}()

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch(component, time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch(component)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().
				Str("component", component).
				Msg("received shutdown signal")
			return
		case <-ticker.C:
			fetcher.Fetch()
			watchdog.Global().Heartbeat(component)
		}
	}
}

func (fetcher *FetcherContractCoordinator) Fetch() {
	component := "FetcherContractCoordinator"

	storage, err := fetcher.coordinatorContract.GetStorage(nil)
	if err != nil {
		logStorageFetchError(component, "coordinator", err)
		return
	}

	// Clear DKG data as it's fetched separately
	storage.Dkg = nil
	storage.PrevDkg = nil

	logFetchSuccess(component, "coordinator")

	jsonData, err := json.Marshal(storage)
	if err != nil {
		logSerializationError(component, "coordinator", err)
		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeContractCoordinator,
		payload: string(jsonData),
	}

	logDataSent(component, "coordinator")
}
