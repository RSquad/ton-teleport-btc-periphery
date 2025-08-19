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
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/httpservice"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
)

type App struct {
	TonClient             *tonclient.TonClient
	BitcoinClient         *bitcoin.Client
	CoordinatorContract   coordinator.Coordinator
	BitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	MetricsService        *metrics.MetricsService
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
		30,
	)

	// Bitcoin client contract
	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		cfg.BitcoinClientContractAddr,
		tonClient,
		nil,
		context.Background(),
	)

	// Open DB connection
	var dbConnPoolMetrics *sql.DB = nil
	{
		dbConnPoolMetrics, err = sql.Open("postgres", cfg.DatabaseUrl)
		if err != nil {
			return nil, err
		}

		// Setup DB pooling (metrics)
		dbConnPoolMetrics.SetMaxOpenConns(cfg.DatabaseMaxConn)
		dbConnPoolMetrics.SetMaxIdleConns(cfg.DatabaseMaxIdleConn)
		dbConnPoolMetrics.SetConnMaxLifetime(-1)
		dbConnPoolMetrics.SetConnMaxIdleTime(-1)
	}

	// Metrics service
	metricsService, err := metrics.NewService(
		coordinatorContract,
		bitcoinClientContract,
		teleportContract,
		bitcoinClient,
		tonClient,
		cfg,
		dbConnPoolMetrics,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create matrics manager: %w", err)
	}

	// HTTP service
	httpService := httpservice.New(
		bitcoinClient,
		tonClient,
		teleportContract,
		dbConnPoolMetrics,
	)

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		TonClient:           tonClient,
		BitcoinClient:       bitcoinClient,
		CoordinatorContract: coordinatorContract,
		MetricsService:      metricsService,
		HttpService:         httpService,
	}, nil
}

func run(app *App) error {
	var wg sync.WaitGroup

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
