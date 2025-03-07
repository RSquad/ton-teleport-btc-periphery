package dkg

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/xssnick/tonutils-go/ton"
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

func (f *Fetcher) Work(ctx context.Context) (err error) {
	logger.DefaultLogStartWork("DKGFetcher")
	defer logger.DefaultLogFinishWork("DKGFetcher", err)
	for {
		dkg, err := f.Fetch(nil)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Fetcher failed")
			time.Sleep(3 * time.Second)
			continue
		}
		f.outChan <- dkg
		time.Sleep(3 * time.Second)
	}
}

func (f *Fetcher) Fetch(block *ton.BlockIDExt) (*coordinator.DKG, error) {
	return f.coordinatorContract.GetDkg(block)
}
