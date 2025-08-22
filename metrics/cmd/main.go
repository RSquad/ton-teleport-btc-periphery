package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/fetchers"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/httpservice"
)

type App struct {
	HttpService    *httpservice.HttpService
	FetcherService *fetchers.FetcherService
	AlertService   *alerts.AlertService
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
		envConfig.BitcoinRpcHost,
		envConfig.BitcoinRpcUser,
		envConfig.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// TON client
	tonClient, err := tonclient.New(envConfig.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Teleport contract
	teleportContract := teleportcontract.New(
		cfg.TeleportContractAddr,
		tonClient,
		nil,
		context.Background(),
	)

	// Coordinator contract
	coordinatorContract := coordinator.New(
		cfg.CoordinatorContractAddr,
		tonClient,
		nil,
		context.Background(),
		30, // TODO: move to config
	)

	// Bitcoin client contract
	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		cfg.BitcoinClientContractAddr,
		tonClient,
		nil,
		context.Background(),
	)

	// Open DB connection
	var dbConnPool *sql.DB = nil
	{
		dbConnPool, err = sql.Open("postgres", cfg.DatabaseUrl)
		if err != nil {
			return nil, err
		}

		// Setup DB pooling (metrics)
		dbConnPool.SetMaxOpenConns(cfg.DatabaseMaxConn)
		dbConnPool.SetMaxIdleConns(cfg.DatabaseMaxIdleConn)
		dbConnPool.SetConnMaxLifetime(-1)
		dbConnPool.SetConnMaxIdleTime(-1)
	}

	// Fetcher service
	fetcherService, err := fetchers.NewService(
		coordinatorContract,
		bitcoinClientContract,
		teleportContract,
		bitcoinClient,
		tonClient,
		cfg,
		dbConnPool,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create matrics manager: %w", err)
	}

	// HTTP service
	httpService := httpservice.New(
		bitcoinClient,
		tonClient,
		teleportContract,
		dbConnPool,
	)

	// Alerts service
	alertsService := alerts.NewAlertService(
		alerts.NewAlertDataSourceLive(dbConnPool, tonClient),
		alerts.NewAlertDispatcherPrometheus(),
		cfg,
	)

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		FetcherService: fetcherService,
		HttpService:    httpService,
		AlertService:   alertsService,
	}, nil
}

func run(app *App) error {
	var wg sync.WaitGroup

	// FetcherService
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.FetcherService.Work(context.Background())
	}()

	// HttpService
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.HttpService.Work(context.Background())
	}()

	// AlertService
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.AlertService.Work(context.Background())
	}()

	wg.Wait()

	log.Println("shutdown complete")
	return nil
}
