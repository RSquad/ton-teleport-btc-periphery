package fetchers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type FetcherDKG struct {
	chDB                chan PayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
	watchdog            *utils.Watchdog
}

func NewFetcherDKG(
	chDB chan PayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
	watchdog *utils.Watchdog,
) *FetcherDKG {
	return &FetcherDKG{
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
		watchdog:            watchdog,
	}
}

func (fetcher *FetcherDKG) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherDKG: stopped")
	logger.DefaultLogStartWork("FetcherDKG: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	fetcher.watchdog.Watch("FetcherDKG")
	defer fetcher.watchdog.Unwatch("FetcherDKG")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.FetchDKG()
			fetcher.FetchPrevDKG()
			fetcher.watchdog.Heartbeat("FetcherDKG")
		}
	}
}

func (fetcher *FetcherDKG) FetchDKG() {
	dkg, err := fetcher.coordinatorContract.GetDkg(nil)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherDKG").
			Msg("fetch failed")

		return
	}

	if dkg == nil {
		logger.Log.Debug().Msg("FetcherDKG: Contract returns dkg==null")
		return
	}

	// Remove DKG R1, R2 packages
	if dkg.R1 != nil {
		dkg.R1.Packages = make(map[uint16][]byte)
	}

	if dkg.R2 != nil {
		dkg.R2.Packages = make(map[uint16][]byte)
	}

	// Serialize
	jsonData, err := json.Marshal(dkg)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("component", "FetcherDKG").
			Msg("failed to serialize DKG->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeDKG,
		payload: string(jsonData),
	}
}

func (fetcher *FetcherDKG) FetchPrevDKG() {
	prevDkg, err := fetcher.coordinatorContract.GetPrevDKG()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("component", "FetcherPrevDKG").
			Msg("fetch failed")

		return
	}

	if prevDkg == nil {
		logger.Log.Debug().Msg("FetcherPrevDKG: Contract returns prevDkg==null")
		return
	}

	// Remove DKG R1, R2 packages
	if prevDkg.R1 != nil {
		prevDkg.R1.Packages = make(map[uint16][]byte)
	}

	if prevDkg.R2 != nil {
		prevDkg.R2.Packages = make(map[uint16][]byte)
	}

	// Serialize
	jsonData, err := json.Marshal(prevDkg)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherPrevDKG").
			Msg("failed to serialize DKG->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypePrevDKG,
		payload: string(jsonData),
	}
}
