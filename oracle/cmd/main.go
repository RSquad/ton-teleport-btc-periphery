package main

import (
	"context"
	"fmt"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/config"
	ton_service "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/ton"
	"github.com/xssnick/tonutils-go/address"
	"log"
)

type App struct {
	Config              config.OracleConfig
	TonClient           *ton.Client
	TonService          *ton_service.TonService
	CoordinatorContract *ton.CoordinatorContract
}

func initialize() (*App, error) {
	var cfg config.OracleConfig

	if err := utils.LoadConfig(&cfg); err != nil {
		log.Fatalf("[App] Failed to load env: %w", err)
	}

	tonClient, err := ton.NewClient()
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	coordinatorContract, err := ton.NewCoordinatorContract(tonClient.API, address.MustParseAddr(cfg.Coordinator), context.Background())
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create wallet contract: %w", err)
	}

	tonService := ton_service.NewTonService(cfg, tonClient, coordinatorContract)

	log.Println("[App] initialized")

	return &App{
		Config:              cfg,
		TonClient:           tonClient,
		TonService:          tonService,
		CoordinatorContract: coordinatorContract,
	}, nil
}

func main() {
	app, err := initialize()
	if err != nil {
		log.Fatalf("[App] failed to initialize: %w", err)
	}

	fmt.Printf("Hello World\n%s", app.Config.TonCenterEndpoint)
}
