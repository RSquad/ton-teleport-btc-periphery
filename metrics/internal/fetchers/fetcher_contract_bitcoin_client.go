package fetchers

import (
	"context"
	"database/sql"
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
	component := "FetcherContractBitcoinClient"

	logger.Log.Info().
		Str("component", component).
		Msg("started")

	defer func() {
		logger.Log.Info().
			Str("component", component).
			Msg("finished")
	}()

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch(component, time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch(component)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().
				Str("component", component).
				Msg("received shutdown signal")
			return
		case <-ticker.C:
			fetcher.Fetch()
			watchdog.Global().Heartbeat(component)
		}
	}
}

func (fetcher *FetcherContractBitcoinClient) Fetch() {
	component := "FetcherContractBitcoinClient"

	storageCell, err := fetcher.bitcoinClientContract.GetStorageCell()
	if err != nil {
		logStorageCellError(component, err)
		return
	}

	// CandidateBlockHashes
	candidateBlockHashes, err := fetcher.bitcoinClientContract.GetCandidateBlockHashesFromCell(storageCell)
	if err != nil {
		logCandidateBlockHashesError(component, err)
		return
	}

	// ConfirmationsNeeded
	confirmationsNeeded := fetcher.bitcoinClientContract.GetConfirmationsNeededFromCell(storageCell)

	// LastConfirmedBlockHash
	lastConfirmedBlockHash, err := fetcher.bitcoinClientContract.GetLastConfirmedBlockHashFromCell(storageCell)
	if err != nil {
		logLastConfirmedBlockHashError(component, err)
		return
	}

	// LastConfirmedBlockHeight
	lastConfirmedBlockHeight, err := fetcher.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		logBlockHeightError(component, lastConfirmedBlockHash.String(), err)
		return
	}

	storage := &data_models.BitcoinClientContractStorage{
		CandidateBlockHashes:     candidateBlockHashes,
		ConfirmationsNeeded:      confirmationsNeeded,
		LastConfirmedBlockHash:   lastConfirmedBlockHash,
		LastConfirmedBlockHeight: lastConfirmedBlockHeight,
	}

	logFetchSuccess(component, "bitcoin_client")

	// Serialize
	jsonData, err := data_models.SerializeBitcoinContractStorageDB(storage)
	if err != nil {
		logSerializationError(component, "bitcoin_client", err)
		return
	}

	fetcher.chDB <- MetricsPayloadDB{
		typeId:  PayloadTypeContractBitcoinClient,
		payload: string(jsonData),
	}

	logDataSent(component, "bitcoin_client")
}
