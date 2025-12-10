package fetchers

import (
	"context"
	"database/sql"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherService struct {
	metricsWriterDB                  *MetricsWriterDB
	eventsWriterDB                   *EventsWriterDB
	fetcherDKG                       *FetcherDKG
	fetcherContractBalances          []*FetcherContractBalance
	fetcherContractBitcoinClient     *FetcherContractBitcoinClient
	fetcherContractTeleport          *FetcherContractTeleport
	fetcherContractCoordinator       *FetcherContractCoordinator
	fetcherEventsContractCoordinator *FetcherEventsContractCoordinator
	fetcherBitcoinNetwork            *FetcherBitcoinNetwork
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
	// Metrics Writer DB
	metricsWriterDbChan := make(chan MetricsPayloadDB, cfg.WriterDbChainSize)
	metricsWriterDB, err := NewMetricsWriterDB(metricsWriterDbChan, db)
	if err != nil {
		return nil, err
	}

	// Events Writer DB
	eventsWriterDbChan := make(chan ton.EventInterface, cfg.WriterDbChainSize)
	eventsWriterDB, err := NewEventsWriterDB(eventsWriterDbChan, db)
	if err != nil {
		return nil, err
	}

	// Fetcher: Contract DKG
	fetcherDKG := NewFetcherDKG(metricsWriterDbChan, coordinatorContract, int64(cfg.DkgFetchPeriod))

	// Fetcher: Contract balances
	fetcherContractBalances := make([]*FetcherContractBalance, 0)
	for name, addr := range contractAddrs {
		fetcherContractBalances = append(fetcherContractBalances,
			NewFetcherContractBalance(db, tonClient, cfg, addr, name),
		)
	}

	// Fetcher: Contract Bitcoin client
	fetcherContractBitcoinClient := NewFetcherContractBitcoinClient(
		metricsWriterDbChan, db, bitcoinClient, bitcoinClientContract, int64(cfg.BitcoinClientContractFetchPeriod))

	// Fetcher: ContractTeleport
	fetcherContractTeleport := NewFetcherContractTeleport(metricsWriterDbChan, teleportContract, int64(cfg.TeleportContractFetchPeriod))

	// Fetcher: ContractCoordinator
	fetcherContractCoordinator := NewFetcherContractCoordinator(metricsWriterDbChan, coordinatorContract, int64(cfg.CoordinatorContractFetchPeriod))

	// Fetcher: Contract Coordinator Events
	fetcherContractCoordinatorEvents := NewFetcherEventsContractCoordinator(
		eventsWriterDbChan,
		coordinator.NewEventParser(),
		tonClient,
		coordinatorContract.GetAddr(),
	)

	// Fetcher: BitcoinNetwork
	fetcherBitcoinNetwork := NewFetcherBitcoinNetwork(metricsWriterDbChan, db, bitcoinClient, int64(cfg.BitcoinNetworkFetchPeriod))

	return &FetcherService{
		metricsWriterDB:                  metricsWriterDB,
		eventsWriterDB:                   eventsWriterDB,
		fetcherDKG:                       fetcherDKG,
		fetcherContractBalances:          fetcherContractBalances,
		fetcherContractBitcoinClient:     fetcherContractBitcoinClient,
		fetcherContractTeleport:          fetcherContractTeleport,
		fetcherContractCoordinator:       fetcherContractCoordinator,
		fetcherEventsContractCoordinator: fetcherContractCoordinatorEvents,
		fetcherBitcoinNetwork:            fetcherBitcoinNetwork,
	}, nil
}

func (s *FetcherService) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("MetricsService: stopped")
	logger.DefaultLogStartWork("MetricsService: starting...")

	var wg sync.WaitGroup

	// Metrics Writer DB
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.metricsWriterDB.Work(ctx)
	}()

	// Events Writer DB
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.eventsWriterDB.Work(ctx)
	}()

	// Fetcher DKG
	if s.fetcherDKG != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.fetcherDKG.Work(ctx)
		}()
	}

	// Fetcher contract balances
	for _, f := range s.fetcherContractBalances {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Work(ctx)
		}()
	}

	// Fetcher ContractBitcoinClient
	if s.fetcherContractBitcoinClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.fetcherContractBitcoinClient.Work(ctx)
		}()
	}

	// Fetcher ContractTeleport
	if s.fetcherContractTeleport != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.fetcherContractTeleport.Work(ctx)
		}()
	}

	// Fetcher ContractCoordinator
	if s.fetcherContractCoordinator != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.fetcherContractCoordinator.Work(ctx)
		}()
	}

	// Fetcher ContractCoordinatorEvents
	if s.fetcherEventsContractCoordinator != nil {
		wg.Add(1)
		go func() {
			s.fetcherEventsContractCoordinator.Work(ctx, &wg)
		}()
	}

	// Fetcher BitcoinNetwork
	if s.fetcherBitcoinNetwork != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.fetcherBitcoinNetwork.Work(ctx)
		}()
	}

	wg.Wait()
}
