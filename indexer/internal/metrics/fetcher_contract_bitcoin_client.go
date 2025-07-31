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

func (fetcher *FetcherContractBitcoinClient) setNextSvbNotZeroMetrics(
	lastPegoutBlockConfirmations int64,
	confirmationsNeeded int64,
	nextSvb uint16,
) {
	nextSvbNotZero.Set(0)
	if lastPegoutBlockConfirmations > confirmationsNeeded {
		if nextSvb != 0 {
			nextSvbNotZero.Set(1)
		}
	}
}

func (fetcher *FetcherContractBitcoinClient) getNextSvbAndLastPegoutHash() (uint16, *chainhash.Hash, error) {
	rows, err := fetcher.db.Query(
		`SELECT payload::json
		FROM metrics_data
		WHERE type_id = 4
		ORDER BY id DESC
		LIMIT 1
	`)
	if err != nil {
		return 0, nil, err
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

	nextSvb, lastPegoutHash, err := fetcher.getNextSvbAndLastPegoutHash()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve NextSVB, error: %v", err))
		return
	}

	lastPegoutBlockHash, err := fetcher.bitcoinClient.GetBlockHashByTxID(lastPegoutHash)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastPegoutBlockHash, error: %v", err))
		return
	}

	lastPegoutHeight, err := fetcher.bitcoinClient.GetBlockHeightByHash(lastPegoutBlockHash)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve LastPegoutHeight, error: %v", err))
		return
	}

	fetcher.setNextSvbNotZeroMetrics(lastConfirmedBlockHeight-lastPegoutHeight, confirmationsNeeded, nextSvb)

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
