package fetchers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
)

type FetcherDKG struct {
	chDB                chan MetricsPayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
}

func NewFetcherDKG(
	chDB chan MetricsPayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
) *FetcherDKG {
	return &FetcherDKG{
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
	}
}

func (fetcher *FetcherDKG) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("FetcherDKG: stopped")
	logger.DefaultLogStartWork("FetcherDKG: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch("FetcherDKG", time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch("FetcherDKG")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.FetchDKG()
			fetcher.FetchPrevDKG()
			watchdog.Global().Heartbeat("FetcherDKG")
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

	fetcher.chDB <- MetricsPayloadDB{
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

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypePrevDKG,
		payload: string(jsonData),
	}
}
