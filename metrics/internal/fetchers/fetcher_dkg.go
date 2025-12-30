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
	component := "FetcherDKG"

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
			fetcher.FetchDKG()
			fetcher.FetchPrevDKG()
			watchdog.Global().Heartbeat(component)
		}
	}
}

func (fetcher *FetcherDKG) FetchDKG() {
	component := "FetcherDKG"

	dkg, err := fetcher.coordinatorContract.GetDkg(nil)
	if err != nil {
		logDkgFetchError(component, err)
		return
	}

	if dkg == nil {
		logNoDkgFound(component)
		return
	}

	// Remove DKG R1, R2 packages
	if dkg.R1 != nil {
		dkg.R1.Packages = make(map[uint16][]byte)
	}

	if dkg.R2 != nil {
		dkg.R2.Packages = make(map[uint16][]byte)
	}

	logFetchSuccess(component, "dkg")

	// Serialize
	jsonData, err := json.Marshal(dkg)
	if err != nil {
		logSerializationError(component, "dkg", err)
		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeDKG,
		payload: string(jsonData),
	}

	logDataSent(component, "dkg")
}

func (fetcher *FetcherDKG) FetchPrevDKG() {
	component := "FetcherPrevDKG"

	prevDkg, err := fetcher.coordinatorContract.GetPrevDKG()
	if err != nil {
		logDkgFetchError(component, err)
		return
	}

	if prevDkg == nil {
		logNoPrevDkgFound(component)
		return
	}

	// Remove DKG R1, R2 packages
	if prevDkg.R1 != nil {
		prevDkg.R1.Packages = make(map[uint16][]byte)
	}

	if prevDkg.R2 != nil {
		prevDkg.R2.Packages = make(map[uint16][]byte)
	}

	logFetchSuccess(component, "previous_dkg")

	// Serialize
	jsonData, err := json.Marshal(prevDkg)
	if err != nil {
		logSerializationError(component, "previous_dkg", err)
		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypePrevDKG,
		payload: string(jsonData),
	}

	logDataSent(component, "previous_dkg")
}
