package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
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
	{
		logMaxSize := helpers.ParseIntWithDefaultVal(cfg.LogMaxSize, 100, "Log max size")
		logMaxBackups := helpers.ParseIntWithDefaultVal(cfg.LogMaxBackups, 50, "Max backups file count")
		logMaxBackupAge := helpers.ParseIntWithDefaultVal(cfg.LogMaxBackupAge, 365, "Max backup file age")

		logLevel, err := logger.ParseLevel(cfg.LogLevel, logger.InfoLevel)
		if err != nil {
			return err
		}

		if err := logger.Init(cfg.LogFile, logLevel, int(logMaxSize), int(logMaxBackups), int(logMaxBackupAge)); err != nil {
			return err
		}
	}

	logger.Log.Info().
		Str("component", "main").
		Msg("Initializing")

	// Keystore
	logger.Log.Info().Msgf("Create a new Keystore at path `%s`", cfg.KeystorePath)
	keystore, err := keystore.New(cfg.KeystorePath)
	if err != nil {
		return err
	}

	// Ton client
	logger.Log.Info().Msgf("Create a new TON client with config `%s`", cfg.TonConfigPathOrURL)
	tonClient, err := tonclient.New(cfg.TonConfigPathOrURL)
	if err != nil {
		return err
	}

	// Validator
	logger.Log.Info().Msg("Create a new Validator")
	validator, err := validator.NewValidator(&cfg)
	if err != nil {
		return err
	}

	// Coordinator address
	logger.Log.Info().Msgf("Parse the Coordinator address `%s`", cfg.CoordinatorContractAddr)
	coordinatorContractAddr, err := address.ParseAddr(cfg.CoordinatorContractAddr)
	if err != nil {
		return err
	}

	// Setup OS signal handlers (SIGINT, SIGTERM)
	logger.Log.Info().Msg("Setup OS signal handler")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancelFn := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	// Coordinator contract
	logger.Log.Info().Msgf("Create a new Coordinator contract wrapper with address `%s`", cfg.CoordinatorContractAddr)
	apiCallTimeout := helpers.ParseIntWithDefaultVal(cfg.ApiCallTimeout, 30, "API call timeout")

	// DKG service
	fetchPeriod := helpers.ParseIntWithDefaultVal(cfg.FetchPeriod, 6, "DKG fetcher period")
	sendStartDKGPeriod := helpers.ParseIntWithDefaultVal(cfg.SendStartDKGPeriod, 10, "SendStartDKG period")
	dkgService := dkg.NewService(
		coordinator.New(coordinatorContractAddr, tonClient, nil, ctx, apiCallTimeout),
		validator,
		fetchPeriod,
		sendStartDKGPeriod,
	)

	// FROST sign service
	executeSignPeriod := helpers.ParseIntWithDefaultVal(cfg.ExecuteSignPeriod, 10, "ExecuteSign period")
	signService := signing.NewService(
		keystore,
		coordinator.New(coordinatorContractAddr, tonClient, nil, ctx, apiCallTimeout),
		tonClient,
		executeSignPeriod,
		&cfg,
	)

	wg.Add(1)
	go dkgService.Work(ctx, &wg, keystore, &cfg)

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
		logger.Log.Info().Msg("All goroutines shut down successfully")
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
