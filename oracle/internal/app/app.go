package app

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/xssnick/tonutils-go/address"
)

type App struct {
	TonClient *tonclient.TonClient
}

func Initialize() (*App, error) {
	logStartInitialization()

	tonClient, err := tonclient.New(cfg.TonConfigUrl)
	if err != nil {
		LogInitializationError(err)
		return nil, err
	}

	coordinatorContractAddr, err := address.ParseAddr(cfg.CoordinatorContractAddr)
	if err != nil {
		LogInitializationError(err)
		return nil, err
	}
	coordinatorContract := coordinator.New(coordinatorContractAddr, tonClient, nil, context.Background())
	dkgService := dkg.NewService(coordinatorContract)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dkgService.Work(context.Background())
	}()

	wg.Wait()

	return &App{
		TonClient: tonClient,
	}, nil
}
