package metrics

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type MetricsService struct {
	writerDB                     *WriterDB
	fetcherDKG                   *FetcherDKG
	fetcherContractBalances      *FetcherContractBalances
	fetcherContractBitcoinClient *FetcherContractBitcoinClient
}

func NewService(
	coordinatorContract *coordinator.CoordinatorContract,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	config config.IndexerConfig,
) (*MetricsService, error) {
	// Writer DB
	writerDbChan := make(chan PayloadDB, 5) // TODO: move 5 to config
	writerDB, err := NewWriterDB(writerDbChan, config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// TODO: reimplement FetcherDKG with all Coordinator data
	// Fetcher: Contract ??? DKG
	fetcherDKG := NewFetcherDKG(writerDbChan, coordinatorContract, 10) // TODO: move 10 to config

	// Fetcher: Contract balances
	fetcherContractBalances := NewFetcherContractBalances(tonClient, config)

	// Fetch: Contract Bitcoin client
	fetcherContractBitcoinClient := NewFetcherContractBitcoinClient(writerDbChan, bitcoinClient, bitcoinClientContract, 60) // TODO: move 60 to config

	return &MetricsService{
		writerDB:                     writerDB,
		fetcherDKG:                   fetcherDKG,
		fetcherContractBalances:      fetcherContractBalances,
		fetcherContractBitcoinClient: fetcherContractBitcoinClient,
	}, nil
}

func (s *MetricsService) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("MetricsService: stopped")
	logger.DefaultLogStartWork("MetricsService: starting...")

	var wg sync.WaitGroup

	// Writer DB
	wg.Add(1)
	go func() {
		s.writerDB.Work(ctx, &wg)
	}()

	// Fetcher DKG
	wg.Add(1)
	go func() {
		s.fetcherDKG.Work(ctx, &wg)
	}()

	// Fetcher contract balances
	wg.Add(1)
	go func() {
		s.fetcherContractBalances.Work(ctx, &wg)
	}()

	// Fetcher ContractBitcoinClient
	wg.Add(1)
	go func() {
		s.fetcherContractBitcoinClient.Work(ctx, &wg)
	}()

	wg.Wait()
}
