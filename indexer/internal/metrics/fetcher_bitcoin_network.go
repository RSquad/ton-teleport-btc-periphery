package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	bu "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/bitcoinutils"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type FetcherBitcoinNetworkData struct {
	Chain         string `json:"chain"`
	Blocks        int32  `json:"blocks"`
	BestBlockHash string `json:"bestblockhash"`
	MedianTime    int64  `json:"mediantime"`
}

type FetcherBitcoinNetwork struct {
	chDB          chan PayloadDB
	db            *sql.DB
	bitcoinClient *bitcoin.Client
	period        int64 // Fetch period (in seconds)
}

type SignedPegout struct {
	createdAt   time.Time
	pegoutAddr  string
	bitcoinTxId string
}

func NewFetcherBitcoinNetwork(
	chDB chan PayloadDB,
	db *sql.DB,
	bitcoinClient *bitcoin.Client,
	period int64,
) *FetcherBitcoinNetwork {
	return &FetcherBitcoinNetwork{
		chDB:          chDB,
		db:            db,
		bitcoinClient: bitcoinClient,
		period:        period,
	}
}

func (fetcher *FetcherBitcoinNetwork) getSignedPegouts() ([]SignedPegout, error) {
	rows, err := fetcher.db.Query(
		`SELECT 
			tt.created_at,
			p.addr AS pegout_addr,
			p.bitcoin_tx_id AS bitcoin_tx_id
		FROM burns AS b
		JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
		JOIN pegouts AS p ON p.id = b.pegout_burn
		WHERE 
			p.status = 'SIGNED'
		-- AND 
			-- created_at > NOW() - INTERVAL '1 day'
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

func (fetcher *FetcherBitcoinNetwork) getLastPegoutId() (*chainhash.Hash, error) {
	rows, err := fetcher.db.Query(
		`	SELECT payload::json
			FROM metrics_data
			WHERE type_id = 4
			ORDER BY id DESC
			LIMIT 1`,
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

func (fetcher FetcherBitcoinNetwork) getCPFPCount(pegoutId *chainhash.Hash) (*bitcoin.TxChildrenCount, error) {
	txChildrenCount, err := fetcher.bitcoinClient.GetTxChildrenCount(pegoutId)
	if err != nil {
		return &bitcoin.TxChildrenCount{}, err
	}
	return txChildrenCount, nil
}

func (fetcher *FetcherBitcoinNetwork) setBitcoinTxExistsMetric(pegouts []SignedPegout) {
	for _, pegout := range pegouts {
		txExists, _, _ := bu.BitcoinTxExists(fetcher.bitcoinClient, pegout.bitcoinTxId)
		if !txExists {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(1)
		} else {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(0)
		}

	}
}

func (fetcher *FetcherBitcoinNetwork) setLastPegoutExistsMetric(pegoutId *chainhash.Hash) error {
	txExists, _, err := bu.BitcoinTxExists(fetcher.bitcoinClient, pegoutId.String())
	if err != nil {
		lastPegout.WithLabelValues(pegoutId.String()).Set(1)
		return err
	}
	if !txExists {
		lastPegout.WithLabelValues(pegoutId.String()).Set(1)
	} else {
		lastPegout.WithLabelValues(pegoutId.String()).Set(0)
	}
	return nil
}

func (fetcher *FetcherBitcoinNetwork) setCPFPCountMetric(count bitcoin.TxChildrenCount) {
	if count.ParentTxID != nil {
		cpfpCounter.WithLabelValues(count.ParentTxID.String()).Set(float64(count.ChildrenCount))
	}
}

func (fetcher *FetcherBitcoinNetwork) Fetch() {
	//
	blockChainInfo, err := fetcher.bitcoinClient.GetBlockChainInfo()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherBitcoinNetwork: failed to retrieve BlockChainInfo, error: %v", err))
		return
	}

	signedPegouts, err := fetcher.getSignedPegouts()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("fetch failed")
	}
	fetcher.setBitcoinTxExistsMetric(signedPegouts)

	LastPegoutTxID, err := fetcher.getLastPegoutId()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("fetch failed")
	}

	err = fetcher.setLastPegoutExistsMetric(LastPegoutTxID)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("fetch failed")
	}

	cpfpCount, err := fetcher.getCPFPCount(LastPegoutTxID)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("fetch failed")
	}
	fetcher.setCPFPCountMetric(*cpfpCount)

	// Serialize
	data := FetcherBitcoinNetworkData{
		Chain:         blockChainInfo.Chain,
		Blocks:        blockChainInfo.Blocks,
		BestBlockHash: blockChainInfo.BestBlockHash,
		MedianTime:    blockChainInfo.MedianTime,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("failed to serialize BlockChainInfo->json")
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypeBlockChainInfo,
		payload:   string(jsonData),
	}
}

func (fetcher *FetcherBitcoinNetwork) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherBitcoinNetwork: stopped")
	logger.DefaultLogStartWork("FetcherBitcoinNetwork: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
