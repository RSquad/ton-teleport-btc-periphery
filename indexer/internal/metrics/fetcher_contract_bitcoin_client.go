package metrics

import (
	"context"
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
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	period                int64 // Fetch period (in seconds)
}

func NewFetcherContractBitcoinClient(
	chDB chan PayloadDB,
	bitcoinClient *bitcoin.Client,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	period int64,
) *FetcherContractBitcoinClient {
	return &FetcherContractBitcoinClient{
		chDB:                  chDB,
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		period:                period,
	}
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

func (fetcher *FetcherContractBitcoinClient) setConfirmedBlockHashMismatchMetric(contractBlockHash string, networkBlockHash string) {
	isMatch := contractBlockHash == networkBlockHash
	if !isMatch {
		confirmedBlockMismatch.WithLabelValues(contractBlockHash, networkBlockHash).Set(1)
		return
	}
	confirmedBlockMismatch.WithLabelValues(contractBlockHash, networkBlockHash).Set(0)
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

	info, err := fetcher.bitcoinClient.GetBlockChainInfo()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBitcoinClient: failed to retrieve BlockChainInfo, error: %v", err))
		return
	}

	// Check if LastConfirmedBlockHashes is match
	fetcher.setConfirmedBlockHashMismatchMetric(lastConfirmedBlockHash.String(), info.BestBlockHash)

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
