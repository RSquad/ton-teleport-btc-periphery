package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type ContractCoordinatorData struct {
	Initiated           bool
	StandaloneMode      bool
	Id                  uint16
	ConfiguratorAddr    string
	Enabled             bool
	Dkg                 *coordinator.DKG
	PrevDkg             *coordinator.DKG
	UnsignedPegouts     []coordinator.PegoutRecord
	PegoutTxCode        *cell.Cell
	MinClaimsPercent    uint16
	MinSignersThreshold uint16
	DkgLifetime         uint32
	SigningTimeout      uint32
	TeleportAddr        string
}

type FetcherContractCoordinator struct {
	chDB                chan PayloadDB
	coordinatorContract *coordinator.CoordinatorContract
	period              int64 // Fetch period (in seconds)
}

func NewFetcherContractCoordinator(
	chDB chan PayloadDB,
	coordinatorContract *coordinator.CoordinatorContract,
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

	contractCoordinatorData := ContractCoordinatorData{
		Initiated:           storage.Initiated,
		StandaloneMode:      storage.StandaloneMode,
		Id:                  storage.Id,
		ConfiguratorAddr:    utils.AddrToRawString(storage.ConfiguratorAddr),
		Enabled:             storage.Enabled,
		Dkg:                 storage.Dkg,
		PrevDkg:             storage.PrevDkg,
		UnsignedPegouts:     storage.UnsignedPegouts,
		PegoutTxCode:        storage.PegoutTxCode,
		MinClaimsPercent:    storage.MinClaimsPercent,
		MinSignersThreshold: storage.MinSignersThreshold,
		DkgLifetime:         storage.DkgLifetime,
		SigningTimeout:      storage.SigningTimeout,
		TeleportAddr:        utils.AddrToRawString(storage.TeleportAddr),
	}

	jsonData, err := json.Marshal(contractCoordinatorData)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractCoordinator").
			Msg("failed to serialize FetcherContractCoordinator->json")
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypeContractCoordinator,
		payload:   string(jsonData),
	}
}
