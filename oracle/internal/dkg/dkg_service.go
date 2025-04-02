package dkg

import (
	"context"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
)

type Service struct {
	coordinatorContract *coordinator.CoordinatorContract
	validator           *validator.Validator
}

func NewService(
	coordinatorContract *coordinator.CoordinatorContract,
	validator *validator.Validator,
) *Service {
	return &Service{
		coordinatorContract: coordinatorContract,
		validator:           validator,
	}
}

func (s *Service) Work(ctx context.Context, wg *sync.WaitGroup, keystore keystore.Keystore) {
	defer wg.Done()
	defer logger.DefaultLogFinishWork("DKGService: started")
	logger.DefaultLogStartWork("DKGService: starting...")

	outChan := make(chan *coordinator.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan, s.coordinatorContract, keystore, s.validator)

	wg.Add(1)
	go fetcher.Work(ctx, wg)

	wg.Add(1)
	go executor.Work(ctx, wg)

	// A periodic event that triggers every 10 seconds to call the SendStartDKG() function
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info().Msg("DKG service received shutdown signal...")
				return
			case _, ok := <-ticker.C:
				if !ok {
					logger.Log.Warn().Msg("Start DKG ticker closed")
					return
				}
				s.coordinatorContract.SendStartDKG()
			}
		}
	}()
}
