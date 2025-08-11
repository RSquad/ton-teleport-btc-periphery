package metrics

import (
	"bytes"
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
	restartCounter      int64
	cfgHash             []byte
	until               time.Time
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
		restartCounter:      0,
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
		fetcher.restartCounter = 0
		fetcher.cfgHash = nil
		fetcher.until = time.Time{}
		dkgStatus.Reset()
		dkgStatus.WithLabelValues("NULL").Set(float64(-1))
		logger.Log.Debug().Msg("FetcherDKG: Contract returns dkg==null")
		return
	}

	if fetcher.until.Equal(time.Time{}) {
		fetcher.until = dkg.Until
		fetcher.cfgHash = dkg.CfgHash
	}

	dkgStatus.Reset()
	dkgStatus.WithLabelValues(dkg.State.String()).Set(float64(dkg.State))
	dkgMaxSigners.Set(float64(dkg.MaxSigners))
	totalValidatorsCount.Set(float64(TOTAL_VALIDATORS))

	if bytes.Equal(fetcher.cfgHash, dkg.CfgHash) &&
		!dkg.Until.Equal(fetcher.until) {
		fetcher.restartCounter++
	}
	//dkgRestartCount.Set(float64(fetcher.restartCounter))

	fetcher.cfgHash = dkg.CfgHash
	fetcher.until = dkg.Until
	storage, err := fetcher.coordinatorContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("component", "FetcherDKG").
			Msg("failed to get coordinator storage")
		return
	}

	threshold := MulDivCeil(uint(storage.MinClaimsPercent), uint(storage.MinSignersThreshold), 100)
	culpritIdx, culpritFound := getCulprit(dkg.Claims.Counters, uint16(threshold))

	culpritRemoved := false
	if culpritFound {
		culpritRemoved = dkg.VSetMask.Bit(int(culpritIdx)) == 0
	}

	setCulpritMetric(culpritFound, culpritRemoved)

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

func setCulpritMetric(culpritFound bool, culpritRemoved bool) {
	dkgCulpritRemains.Set(0)
	dkgCulpritRemoved.Set(0)
	if culpritFound {
		if culpritRemoved {
			dkgCulpritRemoved.Set(1)
		} else {
			dkgCulpritRemains.Set(1)
		}
	}
}

func getCulprit(m map[uint16]uint16, threshold uint16) (uint16, bool) {
	var culprit uint16 = 0
	var claimsCounter uint16 = 0

	if len(m) == 0 {
		return culprit, false
	}

	for i, v := range m {
		if v > claimsCounter {
			claimsCounter = v
			culprit = i
		}
	}

	if claimsCounter < threshold {
		return 0, false
	}

	return culprit, true
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
