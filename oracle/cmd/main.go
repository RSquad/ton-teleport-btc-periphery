package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/wallet"

	coordinatorcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
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

	// jwV4R2Secret, err := hex.DecodeString(oracleConfig.OracleWalletSecret)
	words := strings.Split(oracleConfig.OracleWalletSecret, " ")

	w, err := wallet.FromSeed(tonClient.API, words, wallet.V4R2)
	jwV4R2Secret := w.PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("[App] failed to decode jwv4r2 secret: %w", err)
	}

	jwV4R2Contract, err := jwv4r2contract.NewJWV4R2Contract(
		tonClient.API,
		jwV4R2Secret,
		context.Background(),
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create jwv4r2 contract: %w", err)
	}

	coordinatorContract := coordinatorcontract.New(
		address.MustParseAddr(oracleConfig.CoordinatorContractAddr),
		tonClient,
		jwV4R2Contract,
		context.Background(),
	)

	log.Println("[App] initialized")

	return &App{
		TonClient:           tonClient,
		CoordinatorContract: coordinatorContract,
	}, nil
}

func main() {
	app, err := initialize()
	if err != nil {
		log.Fatalf("[App] failed to initialize: %v", err)
	}

	prevDkg, err := app.CoordinatorContract.GetPrevDKG()
	if err != nil {
		log.Fatalf("[App] Get prev dkg error: %v", err)
	}
	log.Printf("PrevDKG = %v", prevDkg)
}
