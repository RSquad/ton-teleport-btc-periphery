package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/config"
	relayerfactory "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/relayer_factory"
)

type App struct {
	Config         config.RelayerConfig
	TonClient      *tonclient.TonClient
	BitcoinClient  *bitcoin.Client
	JWV4R2Contract *jwv4r2contract.JWV4R2Contract
	RelayerFactory *relayerfactory.RelayerFactory
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
	// Initialize logger
	if err := logger.Init("", logger.DebugLevel, 0, 0, 0); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Log.Info().
		Str("component", "main").
		Msg("initializing")

	// Load configuration
	relayerConfig, err := utils.LoadCfg[config.RelayerConfig]()
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to load configuration")
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize TON client
	tonClient, err := tonclient.New(relayerConfig.TonConfigUrl)
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to initialize TON client")
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	// Initialize Bitcoin client
	bitcoinClient, err := bitcoin.NewClient(
		relayerConfig.BitcoinRpcHost,
		relayerConfig.BitcoinRpcUser,
		relayerConfig.BitcoinRpcPass,
	)
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to initialize Bitcoin client")
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	// Initialize JWV4R2 contract
	jwV4R2Secret, err := hex.DecodeString(relayerConfig.RelayerWallerV4Secret)
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to decode JWV4R2 secret")
		return nil, fmt.Errorf("failed to decode jwv4r2 secret: %w", err)
	}

	jwV4R2Contract, err := jwv4r2contract.NewJWV4R2Contract(
		tonClient.API,
		jwV4R2Secret,
		context.Background(),
	)
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to initialize JWV4R2 contract")
		return nil, fmt.Errorf("failed to create jwv4r2 contract: %w", err)
	}

	logger.Log.Info().
		Str("component", "main").
		Str("relayer_wallet", jwV4R2Contract.Address().StringRaw()).
		Msg("Relayer wallet initialized")

	// Initialize Relayer factory
	relayerFactory := relayerfactory.NewRelayerFactory(bitcoinClient, tonClient)

	logger.Log.Info().
		Str("component", "main").
		Msg("initialized")

	return &App{
		Config:         relayerConfig,
		TonClient:      tonClient,
		BitcoinClient:  bitcoinClient,
		JWV4R2Contract: jwV4R2Contract,
		RelayerFactory: relayerFactory,
	}, nil
}

func run(app *App) error {
	logger.Log.Info().
		Str("component", "main").
		Msg("running")

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start block relayer
	if err := startRelayer(app, "block", 10*time.Second, ctx); err != nil {
		logger.Log.Error().
			Str("component", "main").
			Str("relayer_name", "block").
			Err(err).
			Msg("Failed to start relayer")
		return fmt.Errorf("failed to start block relayer: %w", err)
	}

	// Wait before starting pegout relayer
	time.Sleep(5 * time.Second)

	// Start pegout relayer
	if err := startRelayer(app, "pegout", 10*time.Second, ctx); err != nil {
		logger.Log.Error().
			Str("component", "main").
			Str("relayer_name", "pegout").
			Err(err).
			Msg("Failed to start relayer")
		return fmt.Errorf("failed to start pegout relayer: %w", err)
	}

	// Wait for shutdown signal
	sig := <-sigCh
	logger.Log.Info().
		Str("component", "main").
		Str("signal", sig.String()).
		Msg("Received signal, initiating shutdown")

	cancel()

	logger.Log.Info().
		Str("component", "main").
		Msg("shutdown complete")

	return nil
}

func startRelayer(
	app *App,
	relayerName string,
	interval time.Duration,
	ctx context.Context,
) error {
	component := fmt.Sprintf("Relayer-%s", relayerName)

	logger.Log.Info().
		Str("component", component).
		Dur("interval", interval).
		Msg("Creating relayer")

	relayer, err := app.RelayerFactory.CreateRelayer(
		relayerName,
		app.JWV4R2Contract,
		app.Config.BitcoinClientContractAddr,
		app.Config.TeleportContractAddr,
	)
	if err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("Failed to create relayer")
		return fmt.Errorf("failed to create %s relayer: %w", relayerName, err)
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		logger.Log.Info().
			Str("component", component).
			Msg("Started")

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info().
					Str("component", component).
					Msg("Shutting down")
				return
			case <-ticker.C:
				if err := relayer.Relay(); err != nil {
					logger.Log.Error().
						Str("component", component).
						Err(err).
						Msg("Failed to relay")
				} else {
					logger.Log.Debug().
						Str("component", component).
						Msg("Relay completed successfully")
				}
			}
		}
	}()

	logger.Log.Info().
		Str("component", component).
		Msg("Created successfully")

	return nil
}
