package fetchers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type FetcherContractBitcoinClient struct {
	chDB                  chan PayloadDB
	db                    *sql.DB
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	period                int64 // Fetch period (in seconds)
	watchdog              *utils.Watchdog
}

func NewFetcherContractBitcoinClient(
	chDB chan PayloadDB,
	db *sql.DB,
	bitcoinClient *bitcoin.Client,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	period int64,
	watchdog *utils.Watchdog,
) *FetcherContractBitcoinClient {
	return &FetcherContractBitcoinClient{
		chDB:                  chDB,
		db:                    db,
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		period:                period,
		watchdog:              watchdog,
	}
}

func (fetcher *FetcherContractBitcoinClient) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractBitcoinClient: stopped")
	logger.DefaultLogStartWork("FetcherContractBitcoinClient: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	fetcher.watchdog.Watch("FetcherContractBitcoinClient")
	defer fetcher.watchdog.Unwatch("FetcherContractBitcoinClient")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
			fetcher.watchdog.Heartbeat("FetcherContractBitcoinClient")
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

	storage := &data_models.BitcoinClientContractStorage{
		CandidateBlockHashes:     candidateBlockHashes,
		ConfirmationsNeeded:      confirmationsNeeded,
		LastConfirmedBlockHash:   lastConfirmedBlockHash,
		LastConfirmedBlockHeight: lastConfirmedBlockHeight,
	}

	// Serialize
	jsonData, err := data_models.SerializeBitcoinContractStorageDB(storage)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBitcoinClient").
			Msg("failed to serialize ContractBitcoinClientData->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeContractBitcoinClient,
		payload: string(jsonData),
	}
}
