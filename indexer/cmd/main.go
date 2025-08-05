package main

import (
	"context"
	"database/sql"
	"encoding/hex"
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
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
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

	jwV4R2Secret, err := hex.DecodeString(cfg.ExternalServices.RelayerWalletV4Secret)
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
	var teleportContract *teleportcontract.TeleportContract = nil
	if cfg.ExternalServices.TeleportContractAddr != nil {
		teleportContract = teleportcontract.New(
			cfg.ExternalServices.TeleportContractAddr,
			tonClient,
			nil,
			context.Background(),
		)
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

	// Bitcoin client contract
	var bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract = nil
	if cfg.ExternalServices.BitcoinClientContractAddr != nil {
		bitcoinClientContract = bitcoinclientcontract.NewBitcoinClientContract(
			cfg.ExternalServices.BitcoinClientContractAddr,
			tonClient,
			nil,
			context.Background(),
		)
	}

	repo, err := ent.Open(dialect.Postgres, cfg.ExternalServices.DatabaseUrl)
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
	if cfg.RunServices.RunMintService {
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
	if cfg.RunServices.RunPegoutManager {
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
	if cfg.RunServices.RunEventService {
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
		db, err = sql.Open("postgres", cfg.ExternalServices.DatabaseUrl)
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
	if cfg.RunServices.RunMetricsService {
		metricsService, err = metrics.NewService(
			coordinatorContract,
			bitcoinClientContract,
			teleportContract,
			jwV4R2Contract.Address(),
			bitcoinClient,
			tonClient,
			cfg,
			db,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create matrics manager: %w", err)
		}
	}

	// HTTP service
	var httpService *httpservice.HttpService = nil
	if cfg.RunServices.RunHttpService {
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
