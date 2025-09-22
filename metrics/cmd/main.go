package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

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
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	HttpService    *httpservice.HttpService
	FetcherService *fetchers.FetcherService
	AlertService   *alerts.AlertService
	Watchdog       *utils.Watchdog
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

	// Watchdog
	watchdog, err := utils.NewWatchdog(
		80*time.Second,
		func(id string, overdue time.Duration) {
			logger.Log.Error().Msgf("WATCHDOG: %s missed heartbeat (overdue by %s)", id, overdue)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create watchdog: %w", err)
	}
	logger.Log.Info().Msg("WATCHDOG: created")

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

	// Global runtime config
	globalRuntimeConfig := config.NewGlobalRuntimeConfig(tonClient)

	contractAddrs := map[string]*address.Address{
		"coordinator": cfg.CoordinatorContractAddr,
		"teleport":    cfg.TeleportContractAddr,
		"bitclient":   cfg.BitcoinClientContractAddr,
		"minter":      cfg.JettonMinterContractAddr,
		"relayer":     cfg.RelayerWalletAddr,
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
		contractAddrs,
		watchdog,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetchers: %w", err)
	}

	// Alert manager
	alertManager, err := alerts.NewAlertManager(
		alerts.NewAlertDataSourceLive(dbConnPool, bitcoinClient, globalRuntimeConfig, contractAddrs),
		alerts.NewAlertDispatcherPrometheus(),
		contractAddrs,
		watchdog,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert manager: %w", err)
	}

	// Metrics manager
	metricsManager := metrics.NewMetricsManager(dbConnPool, globalRuntimeConfig, contractAddrs, alertManager)

	// Alerts service
	alertsService := alerts.NewAlertService(
		alertManager,
		cfg,
		watchdog,
	)

	// HTTP service
	httpService := httpservice.New(
		metricsManager,
		alertManager,
		cfg,
	)

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		FetcherService: fetcherService,
		HttpService:    httpService,
		AlertService:   alertsService,
		Watchdog:       watchdog,
	}, nil
}

func run(app *App) error {
	var wg sync.WaitGroup

	// Watchdog
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Watchdog.Start(ctx)
	defer app.Watchdog.Stop()

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
