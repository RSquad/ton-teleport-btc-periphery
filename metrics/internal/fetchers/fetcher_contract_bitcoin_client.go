package fetchers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type FetcherContractBitcoinClient struct {
	chDB                  chan MetricsPayloadDB
	db                    *sql.DB
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	period                int64 // Fetch period (in seconds)
}

func NewFetcherContractBitcoinClient(
	chDB chan MetricsPayloadDB,
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

func (fetcher *FetcherContractBitcoinClient) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("FetcherContractBitcoinClient: stopped")
	logger.DefaultLogStartWork("FetcherContractBitcoinClient: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch("FetcherContractBitcoinClient", time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch("FetcherContractBitcoinClient")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
			watchdog.Global().Heartbeat("FetcherContractBitcoinClient")
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

	// ConfirmationsNeeded
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

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeContractBitcoinClient,
		payload: string(jsonData),
	}
}
