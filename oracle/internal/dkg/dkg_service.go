package dkg

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type Service struct {
	coordinatorContract *coordinatorcontract.CoordinatorContract
}

func NewService(
	coordinatorContract *coordinatorcontract.CoordinatorContract,
) *Service {
	return &Service{
		coordinatorContract: coordinatorContract,
	}
}

func (s *Service) Work(ctx context.Context) (err error) {
	logger.DefaultLogStartWork("DKGService")
	defer logger.DefaultLogFinishWork("DKGService", err)

	outChan := make(chan *coordinatorcontract.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan)

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
