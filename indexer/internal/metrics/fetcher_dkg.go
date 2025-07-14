package metrics

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type FetcherDKG struct {
	chDB                chan PayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
}

func NewFetcherDKG(
	chDB chan PayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
) *FetcherDKG {
	return &FetcherDKG{
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
	}
}

func (fetcher *FetcherDKG) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherDKG: stopped")
	logger.DefaultLogStartWork("FetcherDKG: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.FetchDKG()
			fetcher.FetchPrevDKG()
		}
	}
}

func (fetcher *FetcherDKG) FetchDKG() {
	dkg, err := fetcher.coordinatorContract.GetDkg(nil)
	if err != nil {
		dkgStatus.Reset()
		dkgStatus.WithLabelValues("DKG_ERROR").Set(float64(-1))
		logger.Log.Error().Err(err).
			Str("component", "FetcherDKG").
			Msg("fetch failed")

		return
	}

	if dkg == nil {
		dkgStatus.Reset()
		dkgStatus.WithLabelValues("NULL").Set(float64(-1))
		logger.Log.Debug().Msg("FetcherDKG: Contract returns dkg==null")
		return
	}

	dkgStatus.Reset()
	dkgStatus.WithLabelValues(dkg.State.String()).Set(float64(dkg.State))
	dkgMaxSigners.Set(float64(dkg.MaxSigners))
	totalValidatorsCount.Set(float64(TOTAL_VALIDATORS))

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
		timestamp: time.Now(),
		typeId:    PayloadTypeDKG,
		payload:   string(jsonData),
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

	// Serialize
	jsonData, err := json.Marshal(prevDkg)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherPrevDKG").
			Msg("failed to serialize DKG->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypePrevDKG,
		payload:   string(jsonData),
	}
}
