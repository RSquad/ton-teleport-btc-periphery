package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"entgo.io/ent/dialect"

	entsql "entgo.io/ent/dialect/sql"

	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/migrate"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/events"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/httpservice"
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
	HttpService           *httpservice.HttpService
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
	cfg, err := config.NewServicesConfig(&indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env config: %w", err)
	}

	logger.Log.Debug().Msg(config.CfgToString(&indexerConfig))

	// Bitcoin client
	bitcoinClient, err := bitcoin.NewClient(
		cfg.ExternalServices.BitcoinRpcHost,
		cfg.ExternalServices.BitcoinRpcUser,
		cfg.ExternalServices.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// TON client
	tonClient, err := tonclient.New(cfg.ExternalServices.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Coordinator contract
	var coordinatorContract coordinator.Coordinator = nil
	if cfg.ExternalServices.CoordinatorContractAddr != nil {
		coordinatorContract = coordinator.New(
			cfg.ExternalServices.CoordinatorContractAddr,
			tonClient,
			nil,
			context.Background(),
			30,
		)
	}

	// Teleport contract
	coordinatorContractStorage, err := coordinatorContract.GetStorage(nil)
	if err != nil {
		return nil, fmt.Errorf("FetcherContractCoordinator: failed to retrieve storage cell, error: %v", err)
	}

	teleportContract := teleportcontract.New(
		coordinatorContractStorage.TeleportAddr,
		tonClient,
		nil,
		context.Background(),
	)

	// Open DB connection
	var dbConnPoolGraphql *sql.DB = nil
	{
		dbConnPoolGraphql, err = sql.Open("postgres", cfg.ExternalServices.DatabaseUrl)
		if err != nil {
			return nil, err
		}

		// Setup DB pooling (graphql)
		dbConnPoolGraphql.SetMaxOpenConns(cfg.ExternalServices.DatabaseMaxConn)
		dbConnPoolGraphql.SetMaxIdleConns(cfg.ExternalServices.DatabaseMaxIdleConn)
		dbConnPoolGraphql.SetConnMaxLifetime(1 * time.Minute)
		dbConnPoolGraphql.SetConnMaxIdleTime(1 * time.Minute)
	}

	drv := entsql.OpenDB(dialect.Postgres, dbConnPoolGraphql)
	repo := ent.NewClient(ent.Driver(drv))

	if err := repo.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("failed creating repos schema: %v", err)
	}

	// Mint service
	mintService := mintservice.New(
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
	)

	// Pegout manager
	pegoutManager, err := pegoutmanager.New(
		context.Background(),
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pegout manager: %w", err)
	}

	// Event service
	eventService := events.NewEventService(
		tonClient,
		repo,
		teleportContract,
		coordinatorContract,
	)

	// HTTP service
	httpService := httpservice.New(
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
	)

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
		HttpService:         httpService,
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
