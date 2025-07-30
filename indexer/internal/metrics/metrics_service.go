package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type MetricsService struct {
	writerDB                     *WriterDB
	fetcherDKG                   *FetcherDKG
	fetcherContractBalances      *FetcherContractBalances
	fetcherContractBitcoinClient *FetcherContractBitcoinClient
	fetcherBitcoinNetwork        *FetcherBitcoinNetwork
	fetcherContractTeleport      *FetcherContractTeleport
	fetcherContractCoordinator   *FetcherContractCoordinator
	fetcherPegouts               *FetcherPegouts
}

func NewService(
	coordinatorContract coordinator.Coordinator,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	teleportContract *teleportcontract.TeleportContract,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	cfg *config.ServicesConfig,
	db *sql.DB,
) (*MetricsService, error) {
	// Writer DB
	writerDbChan := make(chan PayloadDB, cfg.Metrics.WriterDbChainSize)
	writerDB, err := NewWriterDB(writerDbChan, db)
	if err != nil {
		return nil, err
	}

	// Fetcher: Contract DKG
	var fetcherDKG *FetcherDKG = nil
	if cfg.Metrics.RunFetcherDKG {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherDKG: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		fetcherDKG = NewFetcherDKG(writerDbChan, coordinatorContract, int64(cfg.Metrics.DkgFetchPeriod))
	}

	// Fetcher: Contract balances
	var fetcherContractBalances *FetcherContractBalances = nil
	if cfg.Metrics.RunFetcherContractBalances {
		fetcherContractBalances = NewFetcherContractBalances(tonClient, cfg)
	}

	// Fetcher: Contract Bitcoin client
	var fetcherContractBitcoinClient *FetcherContractBitcoinClient = nil
	if cfg.Metrics.RunFetcherContractBitcoinClient {
		if bitcoinClientContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractBitcoinClient: BitcoinClientContract is null. Please set the COMMON_TON_CONTRACT_BITCLIENT_ADDR value in the .env")
		}

		fetcherContractBitcoinClient = NewFetcherContractBitcoinClient(writerDbChan, db, bitcoinClient, bitcoinClientContract, int64(cfg.Metrics.BitcoinClientContractFetchPeriod))
	}

	// Fetcher: BitcoinNetwork
	var fetcherBitcoinNetwork *FetcherBitcoinNetwork = nil
	if cfg.Metrics.RunFetcherBitcoinNetwork {
		fetcherBitcoinNetwork = NewFetcherBitcoinNetwork(writerDbChan, bitcoinClient, int64(cfg.Metrics.BitcoinNetworkFetchPeriod))
	}

	// Fetcher: ContractTeleport
	var fetcherContractTeleport *FetcherContractTeleport = nil
	if cfg.Metrics.RunFetcherContractTeleport {
		if teleportContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractTeleport: TeleportContract is null. Please set the COMMON_TON_CONTRACT_TELEPORT_ADDR value in the .env")
		}

		fetcherContractTeleport = NewFetcherContractTeleport(writerDbChan, teleportContract, int64(cfg.Metrics.TeleportContractFetchPeriod))
	}

	// Fetcher: ContractCoordinator
	var fetcherContractCoordinator *FetcherContractCoordinator = nil
	if cfg.Metrics.RunFetcherContractCoordinator {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractCoordinator: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		fetcherContractCoordinator = NewFetcherContractCoordinator(writerDbChan, coordinatorContract, int64(cfg.Metrics.CoordinatorContractFetchPeriod))
	}

	// Fetcher: Pegouts
	var fetcherPegouts *FetcherPegouts = nil
	if cfg.Metrics.RunFetcherPegouts {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherPegouts: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		fetcherPegouts = NewFetcherPegouts(tonClient, bitcoinClient, coordinatorContract, db)
	}

	return &MetricsService{
		writerDB:                     writerDB,
		fetcherDKG:                   fetcherDKG,
		fetcherContractBalances:      fetcherContractBalances,
		fetcherContractBitcoinClient: fetcherContractBitcoinClient,
		fetcherBitcoinNetwork:        fetcherBitcoinNetwork,
		fetcherContractTeleport:      fetcherContractTeleport,
		fetcherContractCoordinator:   fetcherContractCoordinator,
		fetcherPegouts:               fetcherPegouts,
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
	if s.fetcherDKG != nil {
		wg.Add(1)
		go func() {
			s.fetcherDKG.Work(ctx, &wg)
		}()
	}

	// Fetcher contract balances
	if s.fetcherContractBalances != nil {
		wg.Add(1)
		go func() {
			s.fetcherContractBalances.Work(ctx, &wg)
		}()
	}

	// Fetcher ContractBitcoinClient
	if s.fetcherContractBitcoinClient != nil {
		wg.Add(1)
		go func() {
			s.fetcherContractBitcoinClient.Work(ctx, &wg)
		}()
	}

	// Fetcher BitcoinNetwork
	if s.fetcherBitcoinNetwork != nil {
		wg.Add(1)
		go func() {
			s.fetcherBitcoinNetwork.Work(ctx, &wg)
		}()
	}

	// Fetcher ContractTeleport
	if s.fetcherContractTeleport != nil {
		wg.Add(1)
		go func() {
			s.fetcherContractTeleport.Work(ctx, &wg)
		}()
	}

	// Fetcher ContractCoordinator
	if s.fetcherContractCoordinator != nil {
		wg.Add(1)
		go func() {
			s.fetcherContractCoordinator.Work(ctx, &wg)
		}()
	}

	// Fetcher Pegouts
	if s.fetcherPegouts != nil {
		wg.Add(1)
		go func() {
			s.fetcherPegouts.Work(ctx, &wg)
		}()
	}

	wg.Wait()
}
