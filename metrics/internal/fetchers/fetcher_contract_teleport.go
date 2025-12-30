package fetchers

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type FetcherContractTeleport struct {
	chDB             chan MetricsPayloadDB
	teleportContract *teleportcontract.TeleportContract
	period           int64 // Fetch period (in seconds)
}

func NewFetcherContractTeleport(
	chDB chan MetricsPayloadDB,
	teleportContract *teleportcontract.TeleportContract,
	period int64,
) *FetcherContractTeleport {
	return &FetcherContractTeleport{
		chDB:             chDB,
		teleportContract: teleportContract,
		period:           period,
	}
}

func (fetcher *FetcherContractTeleport) Work(ctx context.Context) {
	component := "FetcherContractTeleport"

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

func (fetcher *FetcherContractTeleport) Fetch() {
	component := "FetcherContractTeleport"

	storage, err := fetcher.teleportContract.GetStorage(nil)
	if err != nil {
		logStorageFetchError(component, "teleport", err)
		return
	}

	// Clear code data as it's not needed for metrics
	storage.BlockCode = nil
	storage.PegoutContractCode = nil
	storage.PeginContractCode = nil

	logFetchSuccess(component, "teleport")

	jsonData, err := data_models.SerializeTeleportContractStorage(&storage)
	if err != nil {
		logSerializationError(component, "teleport", err)
		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeContractTeleport,
		payload: string(jsonData),
	}

	logDataSent(component, "teleport")
}
