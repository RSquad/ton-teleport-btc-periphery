package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	bitcoinClient *bitcoin.Client
	period        int64 // Fetch period (in seconds)
}

func NewFetcherBitcoinNetwork(
	chDB chan PayloadDB,
	bitcoinClient *bitcoin.Client,
	period int64,
) *FetcherBitcoinNetwork {
	return &FetcherBitcoinNetwork{
		chDB:          chDB,
		bitcoinClient: bitcoinClient,
		period:        period,
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
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}

func (fetcher *FetcherBitcoinNetwork) Fetch() {
	//
	blockChainInfo, err := fetcher.bitcoinClient.GetBlockChainInfo()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherBitcoinNetwork: failed to retrieve BlockChainInfo, error: %v", err))
		return
	}

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
