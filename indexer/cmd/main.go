package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	"entgo.io/ent/dialect"

	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/migrate"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/events"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/httpservice"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/metrics"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/mintservice"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegoutmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type App struct {
	Repo                  *ent.Client
	TonClient             *tonclient.TonClient
	BitcoinClient         *bitcoin.Client
	EventService          *events.EventService
	TeleportContract      *teleportcontract.TeleportContract
	CoordinatorContract   coordinator.Coordinator
	BitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	PegoutManager         *pegoutmanager.PegoutManager
	MintService           *mintservice.MintService
	MetricsService        *metrics.MetricsService
	HttpService           *httpservice.HttpService
	Db                    *sql.DB
}

func main() {
	log.SetFlags(0)

	app, err := initialize()
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}

	if err := run(app); err != nil {
		log.Fatalf("stopped with error: %v", err)
	}
}

func initialize() (*App, error) {
	if err := logger.Init("", logger.DebugLevel, 0, 0, 0); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Log.Info().
		Str("component", "main").
		Msg("initializing")

	indexerConfig, err := utils.LoadCfg[config.IndexerConfig]()
	if err != nil {
		return nil, err
	}

	// Read .env config
	config, err := config.NewServicesConfig(&indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env config: %w", err)
	}

	// Bitcoin client
	bitcoinClient, err := bitcoin.NewClient(
		config.ExternalServices.BitcoinRpcHost,
		config.ExternalServices.BitcoinRpcUser,
		config.ExternalServices.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// TON client
	tonClient, err := tonclient.New(config.ExternalServices.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Teleport contract
	var teleportContract *teleportcontract.TeleportContract = nil
	if config.ExternalServices.TeleportContractAddr != nil {
		teleportContract = teleportcontract.New(
			config.ExternalServices.TeleportContractAddr,
			tonClient,
			nil,
			context.Background(),
		)
	}

	// Coordinator contract
	var coordinatorContract coordinator.Coordinator = nil
	if config.ExternalServices.CoordinatorContractAddr != nil {
		coordinatorContract = coordinator.New(
			config.ExternalServices.CoordinatorContractAddr,
			tonClient,
			nil,
			context.Background(),
			30,
		)
	}

	// Bitcoin client contract
	var bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract = nil
	if config.ExternalServices.BitcoinClientContractAddr != nil {
		bitcoinClientContract = bitcoinclientcontract.NewBitcoinClientContract(
			config.ExternalServices.BitcoinClientContractAddr,
			tonClient,
			nil,
			context.Background(),
		)
	}

	repo, err := ent.Open(dialect.Postgres, config.ExternalServices.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to create repo: %v", err)
	}

	if err := repo.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("failed creating repos schema: %v", err)
	}

	// Mint service
	var mintService *mintservice.MintService = nil
	if config.RunServices.RunMintService {
		if teleportContract == nil {
			return nil, fmt.Errorf("failed to start MintService: TeleportContract is null. Please set the COMMON_TON_CONTRACT_TELEPORT_ADDR value in the .env")
		}

		mintService = mintservice.New(
			repo,
			bitcoinClient,
			tonClient,
			teleportContract,
		)
	}

	// Pegout manager
	var pegoutManager *pegoutmanager.PegoutManager = nil
	if config.RunServices.RunPegoutManager {
		if teleportContract == nil {
			return nil, fmt.Errorf("failed to start PegoutManager: TeleportContract is null. Please set the COMMON_TON_CONTRACT_TELEPORT_ADDR value in the .env")
		}

		pegoutManager, err = pegoutmanager.New(
			context.Background(),
			repo,
			bitcoinClient,
			tonClient,
			teleportContract,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create pegout manager: %w", err)
		}
	}

	// Event service
	var eventService *events.EventService = nil
	if config.RunServices.RunEventService {
		if teleportContract == nil {
			return nil, fmt.Errorf("failed to start EventService: TeleportContract is null. Please set the COMMON_TON_CONTRACT_TELEPORT_ADDR value in the .env")
		}

		if coordinatorContract == nil {
			return nil, fmt.Errorf("failed to start EventService: CoordinatorContract is null. Please set the COMMON_TON_CONTRACT_COORDINATOR value in the .env")
		}

		eventService = events.NewEventService(
			tonClient,
			repo,
			teleportContract,
			coordinatorContract,
		)
	}

	// Open DB connection
	var db *sql.DB = nil
	{
		db, err = sql.Open("postgres", config.ExternalServices.DatabaseUrl)
		if err != nil {
			return nil, err
		}

		// Setup DB pooling
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(-1)
		db.SetConnMaxIdleTime(-1)
	}

	// Metrics service
	var metricsService *metrics.MetricsService = nil
	if config.RunServices.RunMetricsService {
		metricsService, err = metrics.NewService(
			coordinatorContract,
			bitcoinClientContract,
			teleportContract,
			bitcoinClient,
			tonClient,
			config,
			db,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create matrics manager: %w", err)
		}
	}

	// HTTP service
	var httpService *httpservice.HttpService = nil
	if config.RunServices.RunHttpService {
		httpService = httpservice.New(
			repo,
			bitcoinClient,
			tonClient,
			teleportContract,
			db,
		)
	}

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		Repo:                repo,
		TonClient:           tonClient,
		BitcoinClient:       bitcoinClient,
		TeleportContract:    teleportContract,
		CoordinatorContract: coordinatorContract,
		PegoutManager:       pegoutManager,
		MintService:         mintService,
		EventService:        eventService,
		MetricsService:      metricsService,
		HttpService:         httpService,
		Db:                  db,
	}, nil
}

func run(app *App) error {
	defer app.Repo.Close()

	var wg sync.WaitGroup

	if app.EventService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.EventService.Work(context.Background())
		}()
	}

	if app.PegoutManager != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.PegoutManager.Run()
		}()
	}

	if app.MintService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.MintService.Work(context.Background())
		}()
	}

	if app.MetricsService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.MetricsService.Work(context.Background())
		}()
	}

	if app.HttpService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.HttpService.Work(context.Background())
		}()
	}

	wg.Wait()

	log.Println("shutdown complete")
	return nil
}
