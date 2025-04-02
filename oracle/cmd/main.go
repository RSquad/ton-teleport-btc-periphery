package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

func startAndWaitForStop() error {
	// Load config
	cfg, err := utils.LoadCfg[cfg.Cfg]()
	if err != nil {
		return err
	}

	// Setup logger
	logger.Init(cfg.LogFile)
	logger.Log.Info().
		Str("component", "main").
		Msg("Initializing")

	// Keystore
	logger.Log.Info().Msgf("Trying to create a new Keystore at path `%s`", cfg.KeystorePath)
	keystore, err := keystore.New(cfg.KeystorePath)
	if err != nil {
		return err
	}

	// Ton client
	logger.Log.Info().Msgf("Trying to create a new TON client with config `%s`", cfg.TonConfigPathOrURL)
	tonClient, err := tonclient.New(cfg.TonConfigPathOrURL)
	if err != nil {
		return err
	}

	// Validator
	logger.Log.Info().Msg("Trying to create a new Validator")
	validator, err := validator.NewValidator(&cfg, keystore)
	if err != nil {
		return err
	}

	// Coordinator address
	logger.Log.Info().Msgf("Trying to parse the Coordinator address `%s`", cfg.CoordinatorContractAddr)
	coordinatorContractAddr, err := address.ParseAddr(cfg.CoordinatorContractAddr)
	if err != nil {
		return err
	}

	// Setup OS signal handlers (SIGINT, SIGTERM)
	logger.Log.Info().Msg("Trying to set up OS signal handler")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancelFn := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	// Coordinator contract
	logger.Log.Info().Msgf("Trying to create a new Coordinator contract wrapper with address `%s`", cfg.CoordinatorContractAddr)
	coordinatorContract := coordinator.New(coordinatorContractAddr, tonClient, nil, ctx)

	// DKG service
	dkgService := dkg.NewService(coordinatorContract, validator)

	// FROST sign service
	signService := signing.NewService(keystore, validator, coordinatorContract, tonClient)

	wg.Add(1)
	go dkgService.Work(ctx, &wg, keystore)

	wg.Add(1)
	go signService.Work(ctx, &wg)

	waitForStop(sigChan, cancelFn, &wg)

	return nil
}

func waitForStop(sigChan <-chan os.Signal, cancelFn context.CancelFunc, wg *sync.WaitGroup) {
	// Wait for OS signal
	sig := <-sigChan
	logger.Log.Info().Str("signal", sig.String()).Msg("Received signal")
	logger.Log.Info().Msg("Initiating graceful shutdown...")

	// Cancel the context to notify all goroutines to terminate
	cancelFn()

	// Set a timeout for graceful shutdown
	shutdownChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownChan)
	}()

	// Wait for graceful shutdown with timeout (5 sec)
	select {
	case <-shutdownChan:
		fmt.Println("All goroutines shut down successfully")
	case <-time.After(5 * time.Second):
		logger.Log.Error().Msg("Shutdown timed out, forcing exit")
	}

	logger.Log.Info().Msg("Application stopped")
}

func main() {
	err := startAndWaitForStop()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Str("component", "main").
			Msg("Failed to start")
		return
	}
}
