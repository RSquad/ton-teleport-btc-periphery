package metrics

import (
	"context"
	"sync"
	"time"

	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type FetcherUnsignedPegouts struct {
	tonClient           *tonclient.TonClient
	coordinatorContract coordinator.Coordinator
	expiredAt           map[uint64]time.Time
}

// var expiredAt map[uint64]time.Time

func NewFetcherUnsignedPegouts(tonClient *tonclient.TonClient, coordinator coordinator.Coordinator) *FetcherUnsignedPegouts {
	return &FetcherUnsignedPegouts{
		tonClient:           tonClient,
		coordinatorContract: coordinator,
	}
}

func (f *FetcherUnsignedPegouts) updateUnsignedPegouts(pegouts []coordinator.PegoutRecord) {
	for _, pegout := range pegouts {
		if oldExpiredAt, exists := f.expiredAt[pegout.ID]; exists {
			if oldExpiredAt.Equal(pegout.ExpiredAt) {
				if time.Now().After(pegout.ExpiredAt.Add(PEGOUT_MAX_DELAY)) {
					unsignedPegoutDelayed.WithLabelValues(fmt.Sprint(pegout.ID)).Set(1)
				}
			}
		} else {
			f.expiredAt[pegout.ID] = pegout.ExpiredAt
		}
	}
}

func (f *FetcherUnsignedPegouts) deleteSignedPegouts(pegouts []coordinator.PegoutRecord) {
	for key, _ := range f.expiredAt {
		isSigned := false
		for i, pegout := range pegouts {
			if key == pegout.ID {
				break
			}
			if i == len(pegouts)-1 {
				isSigned = true
			}
		}
		if isSigned {
			delete(f.expiredAt, key)
		}
	}
}

func (f *FetcherUnsignedPegouts) Fetch() {
	unsignedPegouts, err := f.coordinatorContract.GetUnsignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherUnsignedPegouts").
			Msg("fetch failed")
	}

	if unsignedPegouts == nil {
		logger.Log.Debug().Msg("FetcherUnsignedPegouts: Contract returns unsignedPegouts is null")
	}

	if len(unsignedPegouts) == 0 {
		logger.Log.Debug().Msg("FetcherUnsignedPegouts: Contract returns unsignedPegouts is empty")
	}

	unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

	f.updateUnsignedPegouts(unsignedPegouts)
	f.deleteSignedPegouts(unsignedPegouts)
}

func (fetcher *FetcherUnsignedPegouts) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherUnsignedPegouts: stopped")
	logger.DefaultLogStartWork("FetcherUnsignedPegouts: starting...")

	ticker := time.NewTicker(TICKER_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Unsigned Pegouts Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
