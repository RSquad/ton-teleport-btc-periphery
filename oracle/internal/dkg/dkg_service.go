package dkg

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type Service struct {
	coordinatorContract *coordinator.CoordinatorContract
}

func NewService(
	coordinatorContract *coordinator.CoordinatorContract,
) *Service {
	return &Service{
		coordinatorContract: coordinatorContract,
	}
}

func (s *Service) Work(ctx context.Context) (err error) {
	logger.DefaultLogStartWork("DKGService")
	defer logger.DefaultLogFinishWork("DKGService", err)

	outChan := make(chan *coordinator.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan, s.coordinatorContract)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = fetcher.Work(ctx)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Fetcher failed")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = executor.Work(ctx)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Executor failed")
		}
	}()

	wg.Wait()

	return nil
}
