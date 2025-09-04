package fetchers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type FetcherContractTeleport struct {
	chDB             chan PayloadDB
	teleportContract *teleportcontract.TeleportContract
	period           int64 // Fetch period (in seconds)
}

func NewFetcherContractTeleport(
	chDB chan PayloadDB,
	teleportContract *teleportcontract.TeleportContract,
	period int64,
) *FetcherContractTeleport {
	return &FetcherContractTeleport{
		chDB:             chDB,
		teleportContract: teleportContract,
		period:           period,
	}
}

func (fetcher *FetcherContractTeleport) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractTeleport: stopped")
	logger.DefaultLogStartWork("FetcherContractTeleport: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}

func (fetcher *FetcherContractTeleport) Fetch() {
	storage, err := fetcher.teleportContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractTeleport: failed to retrieve storage cell, error: %v", err))
		return
	}

	storage.BlockCode = nil
	storage.PegoutContractCode = nil
	storage.PeginContractCode = nil

	jsonData, err := data_models.SerializeTeleportContractStorage(&storage)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractTeleport: failed to SerializeTeleportContractStorage: %v", err))
		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeContractTeleport,
		payload: string(jsonData),
	}
}
