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
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
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

func (fb *FetcherBitcoinNetwork) getBaseSVB() (float64, error) {
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

func (fb *FetcherBitcoinNetwork) getSVB(blockCount int, mode *btcjson.EstimateSmartFeeMode) (float64, error) {
	if mode == nil {
		mode = &btcjson.EstimateModeEconomical
	}
	fee, err := fb.bitcoinClient.EstimateFee(blockCount, mode)
	if err != nil {
		return 0, err
	}

	return fee, nil
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

func (fb *FetcherBitcoinNetwork) setSVBExceededMetric(svb float64, baseSvb float64) {
	svbCurrent.Set(svb)
	svbBase.Set(baseSvb)
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

	LastPegoutTxID, err := fetcher.getLastPegoutId()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("fetch failed")
	}
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

	baseSvb, err := fetcher.getBaseSVB()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get base svb")
	}
	svb, err := fetcher.getSVB(1, nil)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinTx").
			Msg("fetch failed at get svb")
	}
	fetcher.setSVBExceededMetric(svb, baseSvb)

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

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherBitcoinNetwork received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
