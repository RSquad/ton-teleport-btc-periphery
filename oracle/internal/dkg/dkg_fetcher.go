package dkg

import (
	"context"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type Fetcher struct {
	coordinatorContract *coordinator.CoordinatorContract
	outChan             chan *coordinator.DKG
	period              int64 // Fetch period (in seconds)
}

func NewFetcher(
	coordinatorContract *coordinator.CoordinatorContract,
	outChan chan *coordinator.DKG,
	period int64,
) *Fetcher {
	return &Fetcher{
		coordinatorContract: coordinatorContract,
		outChan:             outChan,
		period:              period,
	}
}

func (f *Fetcher) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.DefaultLogFinishWork("DKG Fetcher")
	logger.DefaultLogStartWork("DKG Fetcher")

	ticker := time.NewTicker(time.Duration(f.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			dkg, err := f.Fetch()
			if err != nil {
				logger.Log.Error().Err(err).
					Str("component", "DKGFetcher").
					Msg("fetch failed")
			} else {
				if dkg != nil {
					f.outChan <- dkg
				}
			}
		}
	}
}

func (f *Fetcher) Fetch() (*coordinator.DKG, error) {
	return f.coordinatorContract.GetDkg(nil)
}
