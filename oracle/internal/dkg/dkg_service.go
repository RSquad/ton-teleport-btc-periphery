package dkg

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
)

type Service struct {
	coordinatorContract *coordinator.CoordinatorContract
	endpoint            *Endpoint
}

func (s *Service) GetClient() *Client {
	return CreateClient(s.endpoint)
}

func NewService(
	coordinatorContract *coordinator.CoordinatorContract,
) *Service {
	return &Service{
		coordinatorContract: coordinatorContract,
		endpoint:            CreateEndpoint(),
	}
}

func (s *Service) Work(ctx context.Context, keystore keystore.Keystore) (err error) {
	logger.DefaultLogStartWork("DKGService")
	defer logger.DefaultLogFinishWork("DKGService", err)

	outChan := make(chan *coordinator.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan, s.coordinatorContract, keystore)

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = executor.startDkgServer(ctx, s.endpoint)
		if err != nil {
			logger.Log.Error().Err(err).Msg("DKG Server failed")
		}
	}()

	wg.Wait()

	return nil
}
