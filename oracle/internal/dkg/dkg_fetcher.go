package dkg

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type Fetcher struct {
	coordinatorContract *coordinator.CoordinatorContract
	outChan             chan *coordinator.DKG
}

func NewFetcher(
	coordinatorContract *coordinator.CoordinatorContract,
	outChan chan *coordinator.DKG,
) *Fetcher {
	return &Fetcher{
		coordinatorContract: coordinatorContract,
		outChan:             outChan,
	}
}

func (f *Fetcher) Work(ctx context.Context) error {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick:
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
