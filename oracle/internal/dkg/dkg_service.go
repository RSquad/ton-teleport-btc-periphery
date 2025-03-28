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

func (s *Service) Work(ctx context.Context, keystore keystore.Keystore) {
	var err error = nil
	logger.DefaultLogStartWork("DKGService")
	defer logger.DefaultLogFinishWork("DKGService", err)

	outChan := make(chan *coordinator.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan, s.coordinatorContract, keystore, s.validator)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.DefaultLogStartWork("DKGFetcher")
		err = fetcher.Work(ctx)
		logger.DefaultLogFinishWork("DKGFetcher", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.DefaultLogStartWork("DKGExecutor")
		err = executor.Work(ctx)
		logger.DefaultLogFinishWork("DKGExecutor", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(10 * time.Second)
		for {
			select {
			case _, ok := <-tick:
				if !ok {
					return
				}
				s.coordinatorContract.SendStartDKG()
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}
