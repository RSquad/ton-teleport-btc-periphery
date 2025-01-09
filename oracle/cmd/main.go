package main

import (
	"log"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/validator"
)

type App struct {
	TonClient *tonclient.TonClient
}

func initialize() (*App, error) {
	cfg, err := utils.LoadCfg[cfg.Cfg]()
	if err != nil {
		return nil, err
	}

	tonClient, err := tonclient.New(cfg.TonConfigUrl)
	if err != nil {
		return nil, err
	}

	_ = validator.New(cfg.Pubkey)

	DGKRoot := dkg.NewRoot(tonClient, &cfg)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		DGKRoot.Runner.Run()
	}()

	wg.Wait()

	return &App{
		TonClient: tonClient,
	}, nil
}

func main() {
	_, err := initialize()
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
}
