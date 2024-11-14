package main

import (
	"context"
	"encoding/hex"
	"fmt"
	libconfig "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/config"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/relayer_factory"
)

type App struct {
	Config         config.RelayerConfig
	TonClient      *ton.Client
	BitcoinClient  *bitcoin.Client
	WalletContract *ton.WalletContract
	RelayerFactory *relayerfactory.RelayerFactory
}

func main() {
	log.SetFlags(0)

	app, err := initialize()
	if err != nil {
		log.Fatalf("[App] failed to initialize: %w", err)
	}

	go startTCPHealthCheck(":3000")

	if err := run(app); err != nil {
		log.Fatalf("[App] stopped with error: %v", err)
	}
}

func initialize() (*App, error) {
	log.Println("[App] initializing...")

	cfg, err := libconfig.LoadConfig[config.RelayerConfig]()
	if err != nil {
		log.Fatalf("[App] Failed to load env: %w", err)
	}

	tonClient, err := ton.NewClient(cfg.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	bitcoinClient, err := bitcoin.NewClient(cfg.BitcoinRpcHost, cfg.BitcoinRpcUser, cfg.BitcoinRpcPass)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create bitcoin client: %w", err)
	}

	walletV4Secret, err := hex.DecodeString(cfg.WalletSecret)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to decode wallet secret: %w", err)
	}

	walletContract, err := ton.NewWalletContract(tonClient.API, walletV4Secret, context.Background())
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create wallet contract: %w", err)
	}

	relayerFactory := relayerfactory.NewRelayerFactory(bitcoinClient, tonClient)

	log.Println("[App] initialized")

	return &App{
		Config:         cfg,
		TonClient:      tonClient,
		BitcoinClient:  bitcoinClient,
		WalletContract: walletContract,
		RelayerFactory: relayerFactory,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := startRelayer(app, "block", app.Config.ContractBitclientAddr, 10*time.Second, ctx); err != nil {
		return fmt.Errorf("[App] failed to start block relayer: %w", err)
	}

	time.Sleep(5 * time.Second)
	if err := startRelayer(app, "pegout", app.Config.ContractTeleportAddr, 20*time.Second, ctx); err != nil {
		return fmt.Errorf("[App] failed to start pegout relayer: %w", err)
	}

	sig := <-sigCh
	log.Printf("[App] received signal: %v. initiating shutdown...", sig)
	cancel()

	log.Println("[App] shutdown complete")
	return nil
}

func startRelayer(app *App, relayerName string, contractAddress string, interval time.Duration, ctx context.Context) error {
	relayer, err := app.RelayerFactory.CreateRelayer(relayerName, app.WalletContract, contractAddress)
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
