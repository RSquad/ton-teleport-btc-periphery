package main

import (
	"context"
	"log"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/config"
)

type App struct {
	TonClient           *tonclient.TonClient
	CoordinatorContract *coordinatorcontract.CoordinatorContract
}

func initialize() (*App, error) {

	oracleConfig, err := utils.LoadConfig[config.OracleConfig]()
	if err != nil {
		return nil, err
	}

	tonClient, err := tonclient.NewTonClient(oracleConfig.TonConfigUrl)
	if err != nil {
		return nil, err
	}

	coordinatorContract, err := coordinatorcontract.NewCoordinatorContract(
		address.MustParseAddr(oracleConfig.CoordinatorContractAddr),
		tonClient,
		context.Background(),
	)
	if err != nil {
		return nil, err
	}

	log.Println("[App] initialized")

	return &App{
		TonClient:           tonClient,
		CoordinatorContract: coordinatorContract,
	}, nil
}

func main() {
	_, err := initialize()
	if err != nil {
		log.Fatalf("[App] failed to initialize: %v", err)
	}
}
