package fetchers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type FetcherBitcoinNetwork struct {
	chDB          chan PayloadDB
	db            *sql.DB
	bitcoinClient *bitcoin.Client
	period        int64 // Fetch period (in seconds)
	watchdog      *utils.Watchdog
}

func NewFetcherBitcoinNetwork(
	chDB chan PayloadDB,
	db *sql.DB,
	bitcoinClient *bitcoin.Client,
	period int64,
	watchdog *utils.Watchdog,
) *FetcherBitcoinNetwork {
	return &FetcherBitcoinNetwork{
		chDB:          chDB,
		db:            db,
		bitcoinClient: bitcoinClient,
		period:        period,
		watchdog:      watchdog,
	}
}

func (fetcher *FetcherBitcoinNetwork) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherBitcoinNetwork: stopped")
	logger.DefaultLogStartWork("FetcherBitcoinNetwork: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	fetcher.watchdog.Watch("FetcherBitcoinNetwork")
	defer fetcher.watchdog.Unwatch("FetcherBitcoinNetwork")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherBitcoinNetwork received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
			fetcher.watchdog.Heartbeat("FetcherBitcoinNetwork")
		}
	}
}

func (fetcher *FetcherBitcoinNetwork) Fetch() {
	blockChainInfo, err := fetcher.bitcoinClient.GetBlockChainInfo()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherBitcoinNetwork: failed to retrieve BlockChainInfo, error: %v", err))
		return
	}

	// Serialize
	bitcoinNetworkInfo := data_models.BitcoinNetworkInfo{
		Chain:         blockChainInfo.Chain,
		Blocks:        blockChainInfo.Blocks,
		BestBlockHash: blockChainInfo.BestBlockHash,
		MedianTime:    blockChainInfo.MedianTime,
	}

	jsonData, err := data_models.SerializeBitcoinNetworkInfoDB(&bitcoinNetworkInfo)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherBitcoinNetwork").
			Msg("failed to serialize BlockChainInfo->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeBitcoinNetwork,
		payload: string(jsonData),
	}
}
