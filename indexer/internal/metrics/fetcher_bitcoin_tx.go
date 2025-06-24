package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	bu "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/bitcoinutils"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

type FetcherBitcoinTx struct {
	bitcoinClient *bitcoin.Client
	db            *sql.DB
	expiredAt     map[uint64]time.Time
}

type SignedPegout struct {
	createdAt   time.Time
	pegoutAddr  string
	bitcoinTxId string
}

func NewFetcherBitcoinTx(
	bitcoinClient *bitcoin.Client,
	db *sql.DB,
) *FetcherBitcoinTx {
	return &FetcherBitcoinTx{
		bitcoinClient: bitcoinClient,
		db:            db,
		expiredAt:     make(map[uint64]time.Time),
	}
}

func (fb *FetcherBitcoinTx) getSignedPegouts() ([]SignedPegout, error) {
	rows, err := fb.db.Query(
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

func (fb *FetcherBitcoinTx) getBaseSVB() (float64, error) {
	rows, err := fb.db.Query(
		`SELECT jsonb_build_object(
				'contractTeleport', (
						SELECT payload::json
						FROM metrics_data
						WHERE type_id = 4
						ORDER BY id DESC
						LIMIT 1
				)
		) AS result;`,
	)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return 0, err
		}
	}

	var storage teleportcontract.Storage
	err = json.Unmarshal([]byte(data), &storage)
	if err != nil {
		return 0, err
	}
	return float64(storage.BaseSVB), nil
}

func (fb *FetcherBitcoinTx) getSVB(blockCount int, mode *btcjson.EstimateSmartFeeMode) (float64, error) {
	if mode == nil {
		mode = &btcjson.EstimateModeEconomical
	}
	fee, err := fb.bitcoinClient.EstimateFee(blockCount, mode)
	if err != nil {
		return 0, err
	}

	return fee, nil
}

func (fb *FetcherBitcoinTx) getLastPegoutId() (*chainhash.Hash, error) {
	rows, err := fb.db.Query(
		`
						SELECT payload::json
						FROM metrics_data
						WHERE type_id = 4
						ORDER BY id DESC
						LIMIT 1
		`,
	)
	if err != nil {
		return &chainhash.Hash{}, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return &chainhash.Hash{}, err
		}
	}

	var storage ContractTeleportData
	err = json.Unmarshal([]byte(data), &storage)
	if err != nil {
		return &chainhash.Hash{}, err
	}
	return storage.LastPegoutTxID, nil
}

func (fb FetcherBitcoinTx) getCPFPCount(pegoutId *chainhash.Hash) (*bitcoin.TxChildrenCount, error) {
	txChildrenCount, err := fb.bitcoinClient.GetTxChildrenCount(pegoutId)
	if err != nil {
		return &bitcoin.TxChildrenCount{}, err
	}
	return txChildrenCount, nil
}

func (fb *FetcherBitcoinTx) setBitcoinTxExistsMetric(pegouts []SignedPegout) {
	for _, pegout := range pegouts {
		txExists, _, _ := bu.BitcoinTxExists(fb.bitcoinClient, pegout.bitcoinTxId)
		if !txExists {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(1)
		} else {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(0)
		}

	}
}

func (fb *FetcherBitcoinTx) setCPFPCountMetric(count bitcoin.TxChildrenCount) {
	if count.ParentTxID != nil {
		cpfpCounter.WithLabelValues(count.ParentTxID.String()).Set(float64(count.ChildrenCount))
	}
}

func (fb *FetcherBitcoinTx) setSVBExceededMetric(exceeded bool) {
	if exceeded {
		svbExceeded.Set(0)
		return
	}
	svbExceeded.Set(1)
}

func (fb *FetcherBitcoinTx) Fetch() {
	fmt.Println("SOME STRING LOG 1")
	signedPegouts, err := fb.getSignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get signed pegouts")
	}

	fb.setBitcoinTxExistsMetric(signedPegouts)
	fmt.Println("SOME STRING LOG 2")
	LastPegoutTxID, err := fb.getLastPegoutId()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get last pegout id")
	}
	fmt.Println("SOME STRING LOG 3")
	cpfpCount, err := fb.getCPFPCount(LastPegoutTxID)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get cpfp count")
	}
	fmt.Println("SOME STRING LOG 4")
	fb.setCPFPCountMetric(*cpfpCount)
	baseSvb, err := fb.getBaseSVB()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get base svb")
	}
	fmt.Println("SOME STRING LOG 5")
	svb, err := fb.getSVB(1, nil)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get svb")
	}
	fmt.Println("svb: ", svb, "baseSvb: ", baseSvb)
	fb.setSVBExceededMetric((svb + SVB_TRESHOLD) > baseSvb)
}

func (fb *FetcherBitcoinTx) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherBitcoinTx: stopped")
	logger.DefaultLogStartWork("FetcherBitcoinTx starting...")

	ticker := time.NewTicker(TICKER_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Bitcoin Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fb.Fetch()
		}
	}
}
