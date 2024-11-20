package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"entgo.io/ent/dialect"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/migrate"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/gql"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/logmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegoutmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	TonCenterV3Client *toncenterv3.Client
	Repo              *ent.Client
	LogManager        *logmanager.LogManager
	PegoutManager     *pegoutmanager.PegoutManager
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

	tonClient, err := tonclient.NewTonClient(indexerConfig.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	teleportContractAddr := address.MustParseAddr(indexerConfig.TeleportContractAddr)
	teleportContract := teleportcontract.New(teleportContractAddr, tonClient.API, nil, context.Background())

	coordinatorContractAddr := address.MustParseAddr(indexerConfig.CoordinatorContractAddr)

	repo, err := ent.Open(dialect.Postgres, indexerConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("[App] failed to create repo: %v", err)
	}

	if err := repo.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		log.Fatalf("[App] failed creating repos schema: %v", err)
	}

	tonCenterV3Client, err := toncenterv3.NewClient(
		indexerConfig.TonCenterV3Host,
		indexerConfig.TonCenterApiKey,
		"/",
		"https",
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	logManager, err := logmanager.New(
		context.Background(),
		repo,
		tonCenterV3Client,
		teleportContract,
		coordinatorContractAddr,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create log manager: %w", err)
	}

	pegoutManager, err := pegoutmanager.New(
		context.Background(),
		repo,
		tonClient,
		tonCenterV3Client,
		teleportContract,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create pegout manager: %w", err)
	}

	log.Println("[App] initialized")

	return &App{
		TonCenterV3Client: tonCenterV3Client,
		Repo:              repo,
		LogManager:        logManager,
		PegoutManager:     pegoutManager,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")
	defer app.Repo.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := handler.NewDefaultServer(gql.NewSchema(app.Repo))
		http.Handle("/",
			playground.ApolloSandboxHandler("Indexer", "/query"),
		)
		http.Handle("/query", srv)
		log.Println("listening on :3001")
		if err := http.ListenAndServe(":3001", nil); err != nil {
			log.Printf("http server terminated: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.LogManager.Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.PegoutManager.Run()
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
