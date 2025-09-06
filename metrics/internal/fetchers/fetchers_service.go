package fetchers

import (
	"context"
	"database/sql"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherService struct {
	writerDB                     *WriterDB
	fetcherDKG                   *FetcherDKG
	fetcherContractBalances      []*FetcherContractBalance
	fetcherContractBitcoinClient *FetcherContractBitcoinClient
	fetcherContractTeleport      *FetcherContractTeleport
	fetcherContractCoordinator   *FetcherContractCoordinator
	fetcherBitcoinNetwork        *FetcherBitcoinNetwork
}

func NewService(
	coordinatorContract coordinator.Coordinator,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	teleportContract *teleportcontract.TeleportContract,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	cfg *config.ServicesConfig,
	db *sql.DB,
	contractAddrs map[string]*address.Address,
) (*FetcherService, error) {
	// Writer DB
	writerDbChan := make(chan PayloadDB, cfg.WriterDbChainSize)
	writerDB, err := NewWriterDB(writerDbChan, db)
	if err != nil {
		return nil, err
	}

	// Fetcher: Contract DKG
	fetcherDKG := NewFetcherDKG(writerDbChan, coordinatorContract, int64(cfg.DkgFetchPeriod))

	// Fetcher: Contract balances
	fetcherContractBalances := make([]*FetcherContractBalance, 0)
	for name, addr := range contractAddrs {
		fetcherContractBalances = append(fetcherContractBalances,
			NewFetcherContractBalance(db, tonClient, cfg, addr, name),
		)
	}

	// Fetcher: Contract Bitcoin client
	fetcherContractBitcoinClient := NewFetcherContractBitcoinClient(writerDbChan, db, bitcoinClient, bitcoinClientContract, int64(cfg.BitcoinClientContractFetchPeriod))

	// Fetcher: ContractTeleport
	fetcherContractTeleport := NewFetcherContractTeleport(writerDbChan, teleportContract, int64(cfg.TeleportContractFetchPeriod))

	// Fetcher: ContractCoordinator
	fetcherContractCoordinator := NewFetcherContractCoordinator(writerDbChan, coordinatorContract, int64(cfg.CoordinatorContractFetchPeriod))

	// Fetcher: BitcoinNetwork
	fetcherBitcoinNetwork := NewFetcherBitcoinNetwork(writerDbChan, db, bitcoinClient, int64(cfg.BitcoinNetworkFetchPeriod))

	return &FetcherService{
		writerDB:                     writerDB,
		fetcherDKG:                   fetcherDKG,
		fetcherContractBalances:      fetcherContractBalances,
		fetcherContractBitcoinClient: fetcherContractBitcoinClient,
		fetcherContractTeleport:      fetcherContractTeleport,
		fetcherContractCoordinator:   fetcherContractCoordinator,
		fetcherBitcoinNetwork:        fetcherBitcoinNetwork,
	}, nil
}

func (s *FetcherService) Work(ctx context.Context) {
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
	for _, f := range s.fetcherContractBalances {
		wg.Add(1)
		go func() {
			f.Work(ctx, &wg)
		}()
	}

	// Fetcher ContractBitcoinClient
	if s.fetcherContractBitcoinClient != nil {
		wg.Add(1)
		go func() {
			s.fetcherContractBitcoinClient.Work(ctx, &wg)
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

	// Fetcher BitcoinNetwork
	if s.fetcherBitcoinNetwork != nil {
		wg.Add(1)
		go func() {
			s.fetcherBitcoinNetwork.Work(ctx, &wg)
		}()
	}

	wg.Wait()
}
