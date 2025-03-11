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
	endpoint            *Endpoint
	validator           *validator.Validator
}

func (s *Service) GetClient() *Client {
	return CreateClient(s.endpoint)
}

func NewService(
	coordinatorContract *coordinator.CoordinatorContract,
	validator *validator.Validator,
) *Service {
	return &Service{
		coordinatorContract: coordinatorContract,
		endpoint:            CreateEndpoint(),
		validator:           validator,
	}
}

func (s *Service) Work(ctx context.Context, keystore keystore.Keystore) (err error) {
	logger.DefaultLogStartWork("DKGService")
	defer logger.DefaultLogFinishWork("DKGService", err)

	outChan := make(chan *coordinator.DKG)
	fetcher := NewFetcher(s.coordinatorContract, outChan)
	executor := NewExecutor(outChan, s.coordinatorContract, keystore, s.validator)

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(10 * time.Second)
		for {
			select {
			case <-tick:
				s.coordinatorContract.SendStartDKG()
			case <-ctx.Done():
				break
			}
		}
	}()

	wg.Wait()

	return nil
}
