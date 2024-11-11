package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/log_listener"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

type App struct {
	TonCenterV3Client           *ton.TonCenterV3Client
	TeleportContractLogListener loglistener.LogListenerInterface
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

	config.LoadEnv()

	tonCenterV3Client, err := ton.NewTonCenterV3Client(false)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	teleportContractAddr := address.MustParseAddr(os.Getenv("COMMON_TON_CONTRACT_TELEPORT_ADDR"))
	teleportContractLogListener, err := loglistener.NewLogListener(tonCenterV3Client, teleportContractAddr)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create teleport contract log listener: %w", err)
	}

	log.Println("[App] initialized")

	return &App{
		TonCenterV3Client:           tonCenterV3Client,
		TeleportContractLogListener: teleportContractLogListener,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)

	go func() {
		done <- app.TeleportContractLogListener.StartListen(ctx)
	}()

	if err := <-done; err != nil {
		log.Printf("[App] LogListener stopped with error: %v", err)
	}

	log.Println("[App] shutdown complete")
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
