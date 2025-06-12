package metrics

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"fmt"

	bu "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/bitcoinutils"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
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
	}
}

func (f *FetcherPegouts) setDelayedMetric(pegouts []coordinator.PegoutRecord) {
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

func (f *FetcherPegouts) getSignedPegouts() ([]SignedPegout, error) {
	rows, err := f.db.Query(
		`SELECT 
			tt.created_at,
			p.addr AS pegout_addr,
			p.bitcoin_tx_id AS bitcoin_tx_id,
			p.status AS status
		FROM burns AS b
		JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
		JOIN pegouts AS p ON p.id = b.pegout_burn
		WHERE 
			status = 'SIGNED'
		AND 
			created_at < NOW() - INTERVAL '1 hour'
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

func (f *FetcherPegouts) checkBitcoinTx(pegouts []SignedPegout) {
	for _, pegout := range pegouts {
		txExists, _, _ := bu.BitcoinTxExists(f.bitcoinClient, pegout.bitcoinTxId)
		if !txExists {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(1)
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
	signedPegouts, err := f.getSignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherPegouts").
			Msg("fetch failed")
	}
	f.checkBitcoinTx(signedPegouts)
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
			logger.Log.Info().Msg("Unsigned Pegouts Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
