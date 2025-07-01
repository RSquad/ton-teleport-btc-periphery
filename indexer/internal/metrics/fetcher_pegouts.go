package metrics

import (
	"context"
	"database/sql"
	"sync"
	"time"

	bu "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/bitcoinutils"
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

type SignedPegout struct {
	createdAt   time.Time
	pegoutAddr  string
	bitcoinTxId string
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
	if len(pegouts) == 0 {
		unsignedPegoutDelayed.WithLabelValues("").Set(0)
		return
	}
	now := time.Now()
	for _, pegout := range pegouts {
		if oldExpiredAt, exists := f.expiredAt[pegout.ID]; exists {
			if oldExpiredAt.Equal(pegout.ExpiredAt) {
				if now.After(pegout.ExpiredAt.Add(PEGOUT_MAX_DELAY)) {
					unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(1)
				} else {
					unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(0)
				}
			}
		} else {
			f.expiredAt[pegout.ID] = pegout.ExpiredAt
		}
	}
}

func (f *FetcherPegouts) deleteSignedPegouts(pegouts []coordinator.PegoutRecord) {
	unsignedPegouts := make(map[uint64]struct{}, len(pegouts))
	for _, pegout := range pegouts {
		unsignedPegouts[pegout.ID] = struct{}{}
	}

	for ID := range f.expiredAt {
		if _, exists := unsignedPegouts[ID]; !exists {
			delete(f.expiredAt, ID) // Delete signed transactions
		}
	}
}

func (f *FetcherPegouts) getSignedPegouts() ([]SignedPegout, error) {
	rows, err := f.db.Query(
		`SELECT 
			tt.created_at,
			p.addr AS pegout_addr,
			p.bitcoin_tx_id AS bitcoin_tx_id
		FROM burns AS b
		JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
		JOIN pegouts AS p ON p.id = b.pegout_burn
		WHERE 
			p.status = 'SIGNED'
		AND 
			created_at > NOW() - INTERVAL '1 day'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return []SignedPegout{}, err
	}

	defer rows.Close()

	var pegouts []SignedPegout
	for rows.Next() {
		var pegout SignedPegout
		err = rows.Scan(&pegout.createdAt, &pegout.pegoutAddr, &pegout.bitcoinTxId)
		if err != nil {
			return []SignedPegout{}, err
		}
		pegouts = append(pegouts, pegout)
	}
	return pegouts, nil
}

func (f *FetcherPegouts) setBitcoinTxExistsMetric(pegouts []SignedPegout) {
	for _, pegout := range pegouts {
		txExists, _, _ := bu.BitcoinTxExists(f.bitcoinClient, pegout.bitcoinTxId)
		if !txExists {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(1)
		} else {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(0)
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

	unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

	f.setDelayedMetric(unsignedPegouts)
	f.deleteSignedPegouts(unsignedPegouts)
	signedPegouts, err := f.getSignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherPegouts").
			Msg("fetch failed")
	}
	f.setBitcoinTxExistsMetric(signedPegouts)
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
