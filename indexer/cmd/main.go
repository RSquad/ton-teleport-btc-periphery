package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"entgo.io/ent/dialect"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/migrate"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/gql"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/logmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegoutmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
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

	bitcoinClient, err := bitcoin.NewClient(
		indexerConfig.BitcoinRpcHost,
		indexerConfig.BitcoinRpcUser,
		indexerConfig.BitcoinRpcPass,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create bitcoin client: %w", err)
	}

	tonClient, err := tonclient.NewTonClient(indexerConfig.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	teleportContractAddr := address.MustParseAddr(indexerConfig.TeleportContractAddr)
	teleportContract := teleportcontract.New(teleportContractAddr, tonClient, nil, context.Background())

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
		bitcoinClient,
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

		mux := http.NewServeMux()
		mux.Handle("/", playground.ApolloSandboxHandler("Indexer", "/graphql"))
		mux.Handle("/graphql", srv)

		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"POST", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})

		handlerWithCORS := c.Handler(mux)

		log.Println("listening on :3000")
		if err := http.ListenAndServe(":3000", handlerWithCORS); err != nil {
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
