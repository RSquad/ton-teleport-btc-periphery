package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
)

type ContractBitcoinClientData struct {
	CandidateBlockHashes     []*chainhash.Hash
	LastConfirmedBlockHash   *chainhash.Hash
	ConfirmationsNeeded      int64
	LastConfirmedBlockHeight int64
}

type FetcherContractBitcoinClient struct {
	chDB                  chan PayloadDB
	db                    *sql.DB
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	period                int64 // Fetch period (in seconds)
}

func NewFetcherContractBitcoinClient(
	chDB chan PayloadDB,
	db *sql.DB,
	bitcoinClient *bitcoin.Client,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	period int64,
) *FetcherContractBitcoinClient {
	return &FetcherContractBitcoinClient{
		chDB:                  chDB,
		db:                    db,
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		period:                period,
	}
}

/*
func (fetcher *FetcherContractBitcoinClient) setNextSvbNotZeroMetrics(
	txId *chainhash.Hash,
	lastPegoutBlockConfirmations int64,
	confirmationsNeeded int64,
	nextSvb uint16,
) {
	nextSvbNotZero.Reset()
	nextSvbNotZero.WithLabelValues(txId.String()).Set(0)
	if lastPegoutBlockConfirmations > confirmationsNeeded {
		if nextSvb != 0 {
			nextSvbNotZero.WithLabelValues(txId.String()).Set(1)
		}
	}
}
*/

/*
func (fetcher *FetcherContractBitcoinClient) getNextSvbAndLastPegoutHash() (uint16, *chainhash.Hash, error) {
	rows, err := fetcher.db.Query(
		`SELECT payload::json
		FROM metrics_data
		WHERE type_id = 4
		ORDER BY id DESC
		LIMIT 1
	`)
	if err != nil {
		return 0, &chainhash.Hash{}, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return 0, &chainhash.Hash{}, err
		}
	}

	var teleportData ContractTeleportData

	err = json.Unmarshal([]byte(data), &teleportData)
	if err != nil {
		return 0, &chainhash.Hash{}, err
	}

	return teleportData.NextSVB, teleportData.LastPegoutTxID, nil
}
*/

func (fetcher *FetcherContractBitcoinClient) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractBitcoinClient: stopped")
	logger.DefaultLogStartWork("FetcherContractBitcoinClient: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
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

/*
func (fetcher *FetcherContractBitcoinClient) setConfirmedBlockHashMismatchMetric(contractBlockHash string, networkBlockHash string) {
	isMatch := contractBlockHash == networkBlockHash
	if !isMatch {
		confirmedBlockMismatch.WithLabelValues(contractBlockHash, networkBlockHash).Set(1)
		return
	}
	confirmedBlockMismatch.WithLabelValues(contractBlockHash, networkBlockHash).Set(0)
}
*/

func (fetcher *FetcherContractBitcoinClient) GetBitcoinInfo() (FetcherBitcoinNetworkData, error) {
	rows, err := fetcher.db.Query(
		`SELECT payload::json
			FROM metrics_data
			WHERE type_id = 3
			ORDER BY id DESC
			LIMIT 1
		`,
	)
	if err != nil {
		return FetcherBitcoinNetworkData{}, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return FetcherBitcoinNetworkData{}, err
		}
	}

	if len(data) == 0 {
		data = "{}"
	}

	var bitcoinNetworkData FetcherBitcoinNetworkData
	err = json.Unmarshal([]byte(data), &bitcoinNetworkData)
	if err != nil {
		return FetcherBitcoinNetworkData{}, err
	}

	return bitcoinNetworkData, nil
}

/*
func (fetcher *FetcherContractBitcoinClient) setDifferentHeightMetric(
	lastBlockHeightClient int64,
	lastBlockHeightNetwork int64,
	confirmationsNeeded int64,
) {
	lastBlockHeightDifference.WithLabelValues(fmt.Sprint(lastBlockHeightNetwork), fmt.Sprint(lastBlockHeightClient)).Set(0)
	if lastBlockHeightNetwork-lastBlockHeightClient > confirmationsNeeded {
		lastBlockHeightDifference.WithLabelValues(fmt.Sprint(lastBlockHeightNetwork), fmt.Sprint(lastBlockHeightClient)).Set(1)
	}
}
*/

func (fetcher *FetcherContractBitcoinClient) Fetch() {
	storageCell, err := fetcher.bitcoinClientContract.GetStorageCell()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve storage cell, error: %v", err))
		return
	}

	// CandidateBlockHashes
	candidateBlockHashes, err := fetcher.bitcoinClientContract.GetCandidateBlockHashesFromCell(storageCell)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve CandidateBlockHashes, error: %v", err))
		return
	}

	// CandidateBlockHashes
	confirmationsNeeded := fetcher.bitcoinClientContract.GetConfirmationsNeededFromCell(storageCell)

	// LastConfirmedBlockHash
	lastConfirmedBlockHash, err := fetcher.bitcoinClientContract.GetLastConfirmedBlockHashFromCell(storageCell)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastConfirmedBlockHash, error: %v", err))
		return
	}

	// LastConfirmedBlockHeight
	lastConfirmedBlockHeight, err := fetcher.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastConfirmedBlockHeight, error: %v", err))
		return
	}

	/*
		"failed to call getblock for hash 0000000497f12786cc8e7236587d2302a9d63fd7d001d85da85c6184aaa29c47: HTTP request failed with status 400 Bad Request, body: {\"error\": \"Unsupported method: / on BITCOIN_SIGNET\"}"
			"failed to call getblock for hash 0000000497f12786cc8e7236587d2302a9d63fd7d001d85da85c6184aaa29c47: HTTP request failed with status 401 Unauthorized, body: Must be authenticated!"
			"failed to call getblock for hash 0000000497f12786cc8e7236587d2302a9d63fd7d001d85da85c6184aaa29c47: HTTP request failed with status 401 Unauthorized, body: Must be authenticated!"
				curl https://bitcoin-rpc.ton-teleport.rsquad.solutions/ \
				-X POST \
				-H "Content-Type: application/json" \
				-d '{
					"jsonrpc": "2.0",
					"id": 1,
					"method": "getblock",
					"params": ["0000000497f12786cc8e7236587d2302a9d63fd7d001d85da85c6184aaa29c47"]
				}'

				curl http://159.223.59.213:38332 \
				-X POST \
				-H "Content-Type: application/json" \
				-d '{
					"jsonrpc": "2.0",
					"id": 1,
					"method": "getblock",
					"params": ["0000000497f12786cc8e7236587d2302a9d63fd7d001d85da85c6184aaa29c47"]
				}'
	*/

	/*
		blockChainInfo, err := fetcher.GetBitcoinInfo()
		if err != nil {
			logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastKnownBlockHeight, error: %v", err))
			return
		}
	*/

	// TODO: check in runtime. The correct comparison should be against the Bitcoin network's block hash at lastConfirmedBlockHeight.
	// Check if LastConfirmedBlockHashes is match
	//	fetcher.setConfirmedBlockHashMismatchMetric(lastConfirmedBlockHash.String(), blockChainInfo.BestBlockHash)

	/*
		nextSvb, lastPegoutHash, err := fetcher.getNextSvbAndLastPegoutHash()
		if err != nil {
			logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve TeleportData, error: %v", err))
			return
		}
	*/

	/*
		lastPegoutBlockHash, err := fetcher.bitcoinClient.GetBlockHashByTxID(lastPegoutHash)
		if err != nil {
			logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastPegoutBlockHash, error: %v", err))
			return
		}
	*/

	/*
		lastPegoutHeight, err := fetcher.bitcoinClient.GetBlockHeightByHash(lastPegoutBlockHash)
		if err != nil {
			logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastPegoutHeight, error: %v", err))
			return
		}
	*/

	//fetcher.setNextSvbNotZeroMetrics(lastPegoutHash, lastConfirmedBlockHeight-lastPegoutHeight, confirmationsNeeded, nextSvb)

	//fetcher.setDifferentHeightMetric(lastConfirmedBlockHeight, int64(blockChainInfo.Blocks), confirmationsNeeded)

	data := &ContractBitcoinClientData{
		CandidateBlockHashes:     candidateBlockHashes,
		ConfirmationsNeeded:      confirmationsNeeded,
		LastConfirmedBlockHash:   lastConfirmedBlockHash,
		LastConfirmedBlockHeight: lastConfirmedBlockHeight,
	}

	// Serialize
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBitcoinClient").
			Msg("failed to serialize ContractBitcoinClientData->json")
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypeContractBitcoinClient,
		payload:   string(jsonData),
	}
}
