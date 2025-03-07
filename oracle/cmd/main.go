package main

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/pegoutsigner"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	TonClient *tonclient.TonClient
}

func initialize() (*App, error) {
	logger.Init()

	logger.Log.Info().
		Str("component", "main").
		Msg("Initializing")

	cfg, err := utils.LoadCfg[cfg.Cfg]()
	if err != nil {
		return nil, err
	}
	keystore, err := keystore.NewKeystore(cfg.KeystorePath)
	if err != nil {
		return nil, err
	}

	tonClient, err := tonclient.New(cfg.TonConfigUrl)
	if err != nil {
		return nil, err
	}

	coordinatorContractAddr, err := address.ParseAddr(cfg.CoordinatorContractAddr)
	if err != nil {
		return nil, err
	}

	coordinatorContract := coordinator.New(coordinatorContractAddr, tonClient, nil, context.Background())
	dkgService := dkg.NewService(coordinatorContract)
	dkgClient := dkgService.GetClient()
	signService := pegoutsigner.NewService(cfg, dkgClient, keystore, nil, coordinatorContract, context.Background())

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dkgService.Work(context.Background())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		signService.Work(context.Background())
	}()

	wg.Wait()

	return &App{
		TonClient: tonClient,
	}, nil
}

func main() {
	_, err := initialize()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("component", "main").
			Msg("Failed to initialize")
		return
	}
}
