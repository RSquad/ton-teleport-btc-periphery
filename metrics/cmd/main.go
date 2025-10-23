package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
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
	Ctx            context.Context
	CancelFn       context.CancelFunc
	Wg             *sync.WaitGroup
	SigChan        <-chan os.Signal
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
	// Setup OS signal handlers (SIGINT, SIGTERM)
	logger.Log.Info().Msg("Setup OS signal handler")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	wg := sync.WaitGroup{}

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
	ctx, cancelFn := context.WithCancel(context.Background())
	err = watchdog.Init(
		10*time.Second,
		60*time.Second,
		60*time.Second,
		func(id string, overdue time.Duration) {
			logger.Log.Error().Str("component", "WATCHDOG").Msgf("'%s' is not responding! Last seen %s seconds ago", id, overdue)
		},
		ctx,
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("failed to initialize watchdog: %w", err)
	}
	logger.Log.Info().Str("component", "WATCHDOG").Msg("Global initialization completed successfully.")

	// Bitcoin client
	bitcoinClient, err := bitcoin.NewClient(
		envConfig.BitcoinRpcHost,
		envConfig.BitcoinRpcUser,
		envConfig.BitcoinRpcPass,
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// TON client
	tonClient, err := tonclient.New(envConfig.TonConfigUrl)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Teleport contract
	teleportContract := teleportcontract.New(
		cfg.TeleportContractAddr,
		tonClient,
		nil,
		ctx,
	)

	// Coordinator contract
	coordinatorContract := coordinator.New(
		cfg.CoordinatorContractAddr,
		tonClient,
		nil,
		ctx,
		30, // TODO: move to config
	)

	// Bitcoin client contract
	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		cfg.BitcoinClientContractAddr,
		tonClient,
		nil,
		ctx,
	)

	// Open DB connection
	var dbConnPool *sql.DB = nil
	{
		dbConnPool, err = sql.Open("postgres", cfg.DatabaseUrl)
		if err != nil {
			cancelFn()
			return nil, err
		}

		// Setup DB pooling (metrics)
		dbConnPool.SetMaxOpenConns(cfg.DatabaseMaxConn)
		dbConnPool.SetMaxIdleConns(cfg.DatabaseMaxIdleConn)
		dbConnPool.SetConnMaxLifetime(-1)
		dbConnPool.SetConnMaxIdleTime(-1)
	}

	// Global runtime config
	config.InitGlobalRuntimeConfig(tonClient, cfg)

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
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("failed to create fetchers: %w", err)
	}

	// Alert manager
	alertManager, err := alerts.NewAlertManager(
		alerts.NewAlertDataSourceLive(dbConnPool, bitcoinClient, contractAddrs),
		alerts.NewAlertDispatcherPrometheus(),
		contractAddrs,
	)
	if err != nil {
		cancelFn()
		return nil, fmt.Errorf("failed to create alert manager: %w", err)
	}

	// Metrics manager
	metricsManager := metrics.NewMetricsManager(
		dbConnPool,
		contractAddrs,
		alertManager,
		bitcoinClient,
	)

	// Alerts service
	alertsService := alerts.NewAlertService(
		alertManager,
		cfg,
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
		Ctx:            ctx,
		CancelFn:       cancelFn,
		Wg:             &wg,
		SigChan:        sigChan,
	}, nil
}

func waitForStop(app *App) {
	// Wait for OS signal
	sig := <-app.SigChan
	logger.Log.Info().Str("signal", sig.String()).Msg("Received signal")
	logger.Log.Info().Msg("Initiating graceful shutdown...")

	// Cancel the context to notify all goroutines to terminate
	app.CancelFn()

	// Set a timeout for graceful shutdown
	shutdownChan := make(chan struct{})
	go func() {
		app.Wg.Wait()
		close(shutdownChan)
	}()

	// Wait for graceful shutdown with timeout (5 sec)
	select {
	case <-shutdownChan:
		logger.Log.Info().Msg("All goroutines shut down successfully")
	case <-time.After(5 * time.Second):
		logger.Log.Error().Msg("Shutdown timed out, forcing exit")
	}

	logger.Log.Info().Msg("Application stopped")
}

func run(app *App) error {
	// FetcherService
	app.Wg.Add(1)
	go func() {
		defer app.Wg.Done()
		app.FetcherService.Work(app.Ctx)
	}()

	// HttpService
	app.Wg.Add(1)
	go func() {
		defer app.Wg.Done()
		app.HttpService.Work(app.Ctx)
	}()

	// AlertService
	app.Wg.Add(1)
	go func() {
		defer app.Wg.Done()
		app.AlertService.Work(app.Ctx)
	}()

	waitForStop(app)

	log.Println("shutdown complete")
	return nil
}
