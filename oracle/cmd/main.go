package main

import (
	"context"
	"fmt"
	libconfig "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/config"
	"github.com/xssnick/tonutils-go/address"
	"log"
)

type App struct {
	Config              config.OracleConfig
	TonClient           *ton.Client
	CoordinatorContract *coordinator_contract.CoordinatorContract
}

func initialize() (*App, error) {

	cfg, err := libconfig.LoadConfig[config.OracleConfig]()
	if err != nil {
		log.Fatalf("[App] Failed to load env: %w", err)
	}

	tonClient, err := ton.NewClient()
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	coordinatorContract, err := coordinator_contract.NewCoordinatorContract(tonClient.API, address.MustParseAddr(cfg.Coordinator), context.Background())
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create wallet contract: %w", err)
	}

	log.Println("[App] initialized")

	return &App{
		Config:              cfg,
		TonClient:           tonClient,
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
