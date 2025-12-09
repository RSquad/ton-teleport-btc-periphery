package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xssnick/tonutils-go/ton/wallet"

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
	PegoutManager         *pegoutmanager.PegoutManager
	MintService           *mintservice.MintService
	HttpService           *httpservice.HttpService
}

func main() {
	log.SetFlags(0)

	if err := logger.Init("", logger.DebugLevel, 0, 0, 0); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	app, err := initialize()
	if err != nil {
		logger.Log.Fatal().
			Str("component", "main").
			Err(err).
			Msg("Failed to initialize")
	}

	if err := run(app); err != nil {
		log.Fatalf("stopped with error: %v", err)
	}
	return repo, nil
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

	highloadWalletV3, err := createTonHighloadWallet(tonClient, cfg.HighLoadWalletV3Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create highload wallet v3: %w", err)
	}

	// Teleport contract
	teleportContract := teleportcontract.New(
		coordinatorContractStorage.TeleportAddr,
		tonClient,
		nil,
		context.Background(),
	)

	repo, err := openDB(cfg.DatabaseUrl, cfg.DatabaseMaxConn, cfg.DatabaseMaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		cfg.BitcoinClientContractAddr,
		tonClient,
		nil,
		context.Background(),
		migrate.WithGlobalUniqueID(true),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		logger.Log.Fatal().
			Str("component", "main").
			Err(err).
			Msg("failed creating repos schema")
	}
	)

	// Mint service
	mintService, err := mintservice.New(
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
		bitcoinClientContract,
		mintservice.NewBatchSender(highloadWalletV3, tonClient),
		highloadWalletV3.Address().String(),
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
	}, nil
}

func createTonHighloadWallet(tonClient *tonclient.TonClient, seed string) (*wallet.Wallet, error) {
	highloadWalletV3, err := wallet.FromSeed(tonClient.API, strings.Split(seed, " "), wallet.ConfigHighloadV3{
		MessageTTL: 60 * 5,
		MessageBuilder: func(ctx context.Context, subWalletId uint32) (id uint32, createdAt int64, err error) {
			// Due to specific of externals emulation on liteserver,
			// we need to take something less than or equals to block time, as message creation time,
			// otherwise external message will be rejected, because time will be > than emulation time
			// hope it will be fixed in the next LS versions
			createdAt = time.Now().Unix() - 30

			// example query id which will allow you to send 1 tx per second
			// but you better to implement your own iterator in database, then you can send unlimited
			// but make sure id is less than 1 << 23, when it is higher start from 0 again
			return uint32(createdAt % (1 << 23)), createdAt, nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create highload wallet v3: %w", err)
	}

	return highloadWalletV3, nil
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
