package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"

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
	CoordinatorContract   coordinator.Coordinator
	BitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	JWV4R2Contract        *jwv4r2contract.JWV4R2Contract
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

func openDB(connectionUrl string, maxOpenConns int, maxIdleConns int) (*ent.Client, error) {
	db, err := sql.Open("postgres", connectionUrl)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	drv := entsql.OpenDB(dialect.Postgres, db)
	repo := ent.NewClient(ent.Driver(drv))

	if err := repo.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		return nil, fmt.Errorf("failed creating repos schema: %w", err)
	}
	return repo, nil
}

func initialize() (*App, error) {
	if err := logger.Init("", logger.DebugLevel, 0, 0, 0); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Log.Info().
		Str("component", "main").
		Msg("initializing")

	envConfig, err := utils.LoadCfg[config.EnvConfig]()
	if err != nil {
		return nil, err
	}

	// Read .env config
	cfg, err := config.NewServicesConfig(&envConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env config: %w", err)
	}

	logger.Log.Debug().Msg(config.CfgToString(cfg))

	// Bitcoin client
	bitcoinClient, err := bitcoin.NewClient(
		cfg.BitcoinRpcHost,
		cfg.BitcoinRpcUser,
		cfg.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// TON client
	tonClient, err := tonclient.New(cfg.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Coordinator contract
	coordinatorContract := coordinator.New(
		cfg.CoordinatorContractAddr,
		tonClient,
		nil,
		context.Background(),
		30,
	)

	coordinatorContractStorage, err := coordinatorContract.GetStorage(nil)
	if err != nil {
		return nil, fmt.Errorf("FetcherContractCoordinator: failed to retrieve storage cell, error: %v", err)
	}

	jwV4R2Secret, err := hex.DecodeString(cfg.IndexerWalletV4Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwv4r2 secret: %w", err)
	}

	jwV4R2Contract, err := jwv4r2contract.NewJWV4R2Contract(
		tonClient.API,
		jwV4R2Secret,
		context.Background(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwv4r2 contract: %w", err)
	}

	// Teleport contract
	teleportContract := teleportcontract.New(
		coordinatorContractStorage.TeleportAddr,
		tonClient,
		jwV4R2Contract,
		context.Background(),
	)

	repo, err := openDB(cfg.DatabaseUrl, cfg.DatabaseMaxConn, cfg.DatabaseMaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		cfg.BitcoinClientContractAddr,
		tonClient,
		jwV4R2Contract,
		context.Background(),
	)

	// Mint service
	mintService, err := mintservice.New(
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
		bitcoinClientContract,
		jwV4R2Contract,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mint service: %w", err)
	}

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
		cfg.PProfHttpEnable,
	)

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		Repo:                  repo,
		TonClient:             tonClient,
		BitcoinClient:         bitcoinClient,
		CoordinatorContract:   coordinatorContract,
		PegoutManager:         pegoutManager,
		MintService:           mintService,
		EventService:          eventService,
		HttpService:           httpService,
		BitcoinClientContract: bitcoinClientContract,
		JWV4R2Contract:        jwV4R2Contract,
	}, nil
}

func run(app *App) error {
	defer app.Repo.Close()

	// Create context that will be cancelled on SIGTERM/SIGINT
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to cancel context when signal is received
	go func() {
		sig := <-sigChan
		logger.Log.Info().
			Str("component", "main").
			Str("signal", sig.String()).
			Msg("received shutdown signal, initiating graceful shutdown")
		cancel()
	}()

	var wg sync.WaitGroup

	if app.EventService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.EventService.Work(ctx)
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
			app.MintService.Work(ctx)
		}()
	}

	if app.HttpService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.HttpService.Work(ctx)
		}()
	}

	wg.Wait()

	logger.Log.Info().
		Str("component", "main").
		Msg("shutdown complete")
	return nil
}
