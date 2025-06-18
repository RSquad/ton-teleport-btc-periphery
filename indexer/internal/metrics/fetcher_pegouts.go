package metrics

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type FetcherPegouts struct {
	tonClient           *tonclient.TonClient
	bitcoinClient       *bitcoin.Client
	coordinatorContract coordinator.Coordinator
	db                  *sql.DB
	expiredAt           map[uint64]time.Time
}

func NewFetcherPegouts(
	tonClient *tonclient.TonClient,
	bitcoinClient *bitcoin.Client,
	coordinator coordinator.Coordinator,
	db *sql.DB,
) *FetcherPegouts {
	return &FetcherPegouts{
		tonClient:           tonClient,
		bitcoinClient:       bitcoinClient,
		db:                  db,
		coordinatorContract: coordinator,
		expiredAt:           make(map[uint64]time.Time),
	}
}

func (f *FetcherPegouts) setDelayedMetric(pegouts []coordinator.PegoutRecord) {
	for _, pegout := range pegouts {
		if oldExpiredAt, exists := f.expiredAt[pegout.ID]; exists {
			if oldExpiredAt.Equal(pegout.ExpiredAt) {
				if time.Now().After(pegout.ExpiredAt.Add(PEGOUT_MAX_DELAY)) {
					unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(1)
				}
			}
		} else {
			f.expiredAt[pegout.ID] = pegout.ExpiredAt
		}
	}
}

func (f *FetcherPegouts) deleteSignedPegouts(pegouts []coordinator.PegoutRecord) {
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

func (f *FetcherPegouts) Fetch() {
	unsignedPegouts, err := f.coordinatorContract.GetUnsignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherPegouts").
			Msg("fetch failed")
	}

	if unsignedPegouts == nil {
		logger.Log.Debug().Msg("FetcherPegouts: Contract returns unsignedPegouts is null")
	}

	if len(unsignedPegouts) == 0 {
		logger.Log.Debug().Msg("FetcherPegouts: Contract returns unsignedPegouts is empty")
	}

	unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

	f.setDelayedMetric(unsignedPegouts)
	f.deleteSignedPegouts(unsignedPegouts)
}

func (fetcher *FetcherPegouts) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherPegouts: stopped")
	logger.DefaultLogStartWork("FetcherPegouts: starting...")

	ticker := time.NewTicker(TICKER_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Pegouts Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
