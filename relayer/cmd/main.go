package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
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
		log.Fatalf("[App] failed to initialize: %v", err)
	}

	go startTCPHealthCheck(":3000")

	if err := run(app); err != nil {
		log.Fatalf("[App] stopped with error: %v", err)
	}
}

func initialize() (*App, error) {
	log.Println("[App] initializing...")

	relayerConfig, err := utils.LoadCfg[config.RelayerConfig]()
	if err != nil {
		log.Fatalf("[App] Failed to load env: %v", err)
	}

	tonClient, err := tonclient.New(relayerConfig.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	bitcoinClient, err := bitcoin.NewClient(
		relayerConfig.BitcoinRpcHost,
		relayerConfig.BitcoinRpcUser,
		relayerConfig.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create bitcoin client: %w", err)
	}

	jwV4R2Secret, err := hex.DecodeString(relayerConfig.RelayerWallerV4Secret)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to decode jwv4r2 secret: %w", err)
	}

	jwV4R2Contract, err := jwv4r2contract.NewJWV4R2Contract(
		tonClient.API,
		jwV4R2Secret,
		context.Background(),
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create jwv4r2 contract: %w", err)
	}

	log.Printf("[App] Relayer wallet: %s", jwV4R2Contract.Address().StringRaw())

	relayerFactory := relayerfactory.NewRelayerFactory(bitcoinClient, tonClient)

	log.Println("[App] initialized")

	return &App{
		Config: relayerConfig, TonClient: tonClient, BitcoinClient: bitcoinClient, JWV4R2Contract: jwV4R2Contract, RelayerFactory: relayerFactory,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := startRelayer(app, "block", 10*time.Second, ctx); err != nil {
		return fmt.Errorf("[App] failed to start block relayer: %w", err)
	}

	time.Sleep(5 * time.Second)
	if err := startRelayer(app, "pegout", 10*time.Second, ctx); err != nil {
		return fmt.Errorf("[App] failed to start pegout relayer: %w", err)
	}

	sig := <-sigCh
	log.Printf("[App] received signal: %v. initiating shutdown...", sig)
	cancel()

	log.Println("[App] shutdown complete")
	return nil
}

func startRelayer(
	app *App,
	relayerName string,
	interval time.Duration,
	ctx context.Context,
) error {
	relayer, err := app.RelayerFactory.CreateRelayer(
		relayerName,
		app.JWV4R2Contract,
		app.Config.BitcoinClientContractAddr,
		app.Config.TeleportContractAddr,
	)
	if err != nil {
		return fmt.Errorf("[App] failed to create %v relayer: %w", relayerName, err)
	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		log.Printf("[App] %v relayer started", relayerName)
		for {
			select {
			case <-ctx.Done():
				log.Printf("[App] shutting down %v relayer", relayerName)
				return
			case <-ticker.C:
				if err := relayer.Relay(); err != nil {
					log.Printf("[App] failed to relay %v: %v", relayerName, err)
				}
			}
		}
	}()

	return nil
}

func startTCPHealthCheck(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("[App] failed to start healthcheck server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[App] failed to accept healthcheck ping: %v", err)
			continue
		}

		log.Println("[App] healthcheck pong")
		conn.Close()
	}
}
