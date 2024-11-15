package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"entgo.io/ent/dialect"

	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/ent"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/logmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	TonCenterV3Client *ton.TonCenterV3Client
	Repo              *ent.Client
	LogManager        *logmanager.LogManager
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

	indexerConfig, err := utils.LoadConfig[config.IndexerConfig]()
	if err != nil {
		return nil, err
	}

	teleportContractAddr := address.MustParseAddr(indexerConfig.TeleportContractAddr)

	repo, err := ent.Open(dialect.Postgres, indexerConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("[App] failed to create repo: %v", err)
	}
	defer repo.Close()

	if err := repo.Schema.Create(context.Background()); err != nil {
		log.Fatalf("[App] failed creating repos schema: %v", err)
	}

	tonCenterV3Client, err := ton.NewTonCenterV3Client(
		indexerConfig.TonCenterV3Host,
		indexerConfig.TonCenterApiKey,
		"/",
		"https",
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	logManager, err := logmanager.New(tonCenterV3Client, teleportContractAddr)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create log manager: %w", err)
	}

	log.Println("[App] initialized")

	return &App{
		TonCenterV3Client: tonCenterV3Client,
		Repo:              repo,
		LogManager:        logManager,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.LogManager.Run()
	}()

	wg.Wait()

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
