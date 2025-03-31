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
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/signing"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
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
	keystore, err := keystore.New(cfg.KeystorePath)
	if err != nil {
		return nil, err
	}

	tonClient, err := tonclient.New(cfg.TonConfigPathOrURL)
	if err != nil {
		return nil, err
	}

	validator, err := validator.NewValidator(&cfg, keystore)
	if err != nil {
		return nil, err
	}

	coordinatorContractAddr, err := address.ParseAddr(cfg.CoordinatorContractAddr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coordinatorContract := coordinator.New(coordinatorContractAddr, tonClient, nil, ctx)
	dkgService := dkg.NewService(coordinatorContract, validator)
	signService := signing.NewService(keystore, validator, coordinatorContract, tonClient)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dkgService.Work(ctx, keystore)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		signService.Work(ctx)
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
