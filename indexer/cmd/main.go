package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/migrate"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/events"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/gql"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/mintmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegoutmanager"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	Repo                *ent.Client
	TonClient           *tonclient.TonClient
	BitcoinClient       *bitcoin.Client
	EventService        *events.EventService
	TeleportContract    *teleportcontract.TeleportContract
	CoordinatorContract *coordinatorcontract.CoordinatorContract
	PegoutManager       *pegoutmanager.PegoutManager
	MintManager         *mintmanager.MintManager
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
	logger.Init()

	logger.Log.Info().
		Str("component", "main").
		Msg("initializing")

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
		return nil, fmt.Errorf("failed to create bitcoin client: %w", err)
	}

	tonClient, err := tonclient.New(indexerConfig.TonConfigUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ton client: %w", err)
	}

	teleportContractAddr := address.MustParseAddr(indexerConfig.TeleportContractAddr)
	teleportContract := teleportcontract.New(
		teleportContractAddr,
		tonClient,
		nil,
		context.Background(),
	)

	coordinatorContractAddr := address.MustParseAddr(indexerConfig.CoordinatorContractAddr)
	coordinatorContract := coordinatorcontract.New(
		nil,
		coordinatorContractAddr,
		tonClient,
		context.Background(),
	)

	repo, err := ent.Open(dialect.Postgres, indexerConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create repo: %v", err)
	}

	if err := repo.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		log.Fatalf("failed creating repos schema: %v", err)
	}

	mintManager := mintmanager.New(
		context.Background(),
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
	)

	pegoutManager, err := pegoutmanager.New(
		context.Background(),
		repo,
		bitcoinClient,
		tonClient,
		teleportContract,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pegout manager: %w", err)
	}

	eventService := events.NewEventService(
		tonClient,
		repo,
		teleportContract,
		coordinatorContract,
	)

	log.Println("initialized")

	return &App{
		Repo:                repo,
		TonClient:           tonClient,
		BitcoinClient:       bitcoinClient,
		TeleportContract:    teleportContract,
		CoordinatorContract: coordinatorContract,
		PegoutManager:       pegoutManager,
		MintManager:         mintManager,
		EventService:        eventService,
	}, nil
}

func run(app *App) error {
	defer app.Repo.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := handler.NewDefaultServer(
			gql.NewSchema(app.Repo, app.BitcoinClient, app.TeleportContract),
		)
		srv.Use(entgql.Transactioner{TxOpener: app.Repo})

		mux := http.NewServeMux()
		mux.Handle("/indexer/graphql", srv)
		mux.Handle("/", playground.ApolloSandboxHandler("Indexer", "/indexer/graphql"))

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
		app.EventService.Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.PegoutManager.Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.MintManager.Run()
	}()

	wg.Wait()

	log.Println("shutdown complete")
	return nil
}
