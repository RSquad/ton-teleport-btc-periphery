package main

import (
	"context"
	"log"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/config"
	"github.com/xssnick/tonutils-go/address"
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

	tonClient, err := tonclient.New(oracleConfig.TonConfigUrl)
	if err != nil {
		return nil, err
	}

	coordinatorContract := coordinatorcontract.New(
		signer.New(),
		address.MustParseAddr(oracleConfig.CoordinatorContractAddr),
		tonClient,
		context.Background(),
	)

	return &App{
		TonClient:           tonClient,
		CoordinatorContract: coordinatorContract,
	}, nil
}

func main() {
	app, err := initialize()
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}

	block, err := app.TonClient.API.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalf("failed to get current masterchain info: %v", err)
	}

	dkg, err := app.CoordinatorContract.GetDkg(block)
	if err != nil {
		log.Fatalf("failed to get dkg: %v", err)
	}

	log.Printf("dkg: %+v", dkg)

	tx, err := app.CoordinatorContract.SendStartDKG(0)
	if err != nil {
		log.Fatalf("failed to send start dkg: %v", err)
	}

	log.Printf("tx: %+v", tx)
}
