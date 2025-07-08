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
	cfg config.IndexerConfig,
	db *sql.DB,
) (*MetricsService, error) {
	// Writer DB
	dbChainSize, err := config.ParseIntWithDefaultVal(cfg.MetricsWriterDbChainSize, 5, "MetricsWriterDbChainSize")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	writerDbChan := make(chan PayloadDB, dbChainSize)
	writerDB, err := NewWriterDB(writerDbChan, db)
	if err != nil {
		return nil, err
	}

	// Fetcher: Contract DKG
	runMetricsFetcherDKG, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherDKG, true, "RunMetricsFetcherDKG")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherDKG *FetcherDKG = nil
	if runMetricsFetcherDKG {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherDKG: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		dkgFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsDkgFetchPeriod, 10, "MetricsDkgFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherDKG = NewFetcherDKG(writerDbChan, coordinatorContract, dkgFetchPeriod)
	}

	// Fetcher: Contract balances
	runMetricsFetcherContractBalances, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherContractBalances, true, "RunMetricsFetcherContractBalances")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherContractBalances *FetcherContractBalances = nil
	if runMetricsFetcherContractBalances {
		fetcherContractBalances = NewFetcherContractBalances(tonClient, cfg)
	}

	// Fetcher: Contract Bitcoin client
	runMetricsFetcherContractBitcoinClient, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherContractBitcoinClient, true, "RunMetricsFetcherContractBitcoinClient")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherContractBitcoinClient *FetcherContractBitcoinClient = nil
	if runMetricsFetcherContractBitcoinClient {
		if bitcoinClientContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractBitcoinClient: BitcoinClientContract is null. Please set the COMMON_TON_CONTRACT_BITCLIENT_ADDR value in the .env")
		}

		bitcoinClientContractFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsBitcoinClientContractFetchPeriod, 60, "MetricsBitcoinClientContractFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherContractBitcoinClient = NewFetcherContractBitcoinClient(writerDbChan, bitcoinClient, bitcoinClientContract, bitcoinClientContractFetchPeriod)
	}

	// Fetcher: BitcoinNetwork
	runMetricsFetcherBitcoinNetwork, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherBitcoinNetwork, true, "RunMetricsFetcherBitcoinNetwork")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherBitcoinNetwork *FetcherBitcoinNetwork = nil
	if runMetricsFetcherBitcoinNetwork {
		bitcoinNetworkFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsBitcoinNetworkFetchPeriod, 59, "MetricsBitcoinNetworkFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherBitcoinNetwork = NewFetcherBitcoinNetwork(writerDbChan, db, bitcoinClient, bitcoinNetworkFetchPeriod)
	}

	// Fetcher: ContractTeleport
	runMetricsFetcherContractTeleport, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherContractTeleport, true, "RunMetricsFetcherContractTeleport")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherContractTeleport *FetcherContractTeleport = nil
	if runMetricsFetcherContractTeleport {
		if teleportContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractTeleport: TeleportContract is null. Please set the COMMON_TON_CONTRACT_TELEPORT_ADDR value in the .env")
		}

		teleportContractFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsTeleportContractFetchPeriod, 27, "MetricsTeleportContractFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherContractTeleport = NewFetcherContractTeleport(writerDbChan, teleportContract, teleportContractFetchPeriod)
	}

	// Fetcher: ContractCoordinator
	runMetricsFetcherContractCoordinator, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherContractCoordinator, true, "RunMetricsFetcherContractCoordinator")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherContractCoordinator *FetcherContractCoordinator = nil
	if runMetricsFetcherContractCoordinator {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherContractCoordinator: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		coordinatorContractFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsCoordinatorContractFetchPeriod, 59, "MetricsCoordinatorContractFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherContractCoordinator = NewFetcherContractCoordinator(writerDbChan, coordinatorContract, coordinatorContractFetchPeriod)
	}

	// Fetcher: Pegouts
	runMetricsFetcherPegouts, err := config.ParseBoolWithDefaultVal(cfg.RunMetricsFetcherPegouts, true, "RunMetricsFetcherPegouts")
	if err != nil {
		logger.Log.Error().Str("component", "metrics").Err(err)
		return nil, err
	}

	var fetcherPegouts *FetcherPegouts = nil
	if runMetricsFetcherPegouts {
		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start FetcherPegouts: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		pegoutMetricsFetchPeriod, err := config.ParseIntWithDefaultVal(cfg.MetricsPegoutsFetchPeriod, 61, "MetricsPegoutsFetchPeriod")
		if err != nil {
			logger.Log.Error().Str("component", "metrics").Err(err)
			return nil, err
		}

		fetcherPegouts = NewFetcherPegouts(tonClient, bitcoinClient, coordinatorContract, db, pegoutMetricsFetchPeriod)
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

func (s *MetricsService) Work(ctx context.Context, indexerConfig *config.IndexerConfig) {
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

	// Fetcher BitcoinTx
	wg.Add(1)
	go func() {
		s.fetcherBitcoinTx.Work(ctx, &wg)
	}()

	wg.Wait()
}
