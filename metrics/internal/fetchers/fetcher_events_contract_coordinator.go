package fetchers

import (
	"context"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherEventsContractCoordinator struct {
	chDB               chan ton.EventInterface
	parser             ton.EventParserInterface
	tonClient          *tonclient.TonClient
	coordinatorAddress *address.Address
}

func NewFetcherEventsContractCoordinator(
	chDB chan ton.EventInterface,
	parser ton.EventParserInterface,
	tonClient *tonclient.TonClient,
	coordinatorAddress *address.Address,
) *FetcherEventsContractCoordinator {
	return &FetcherEventsContractCoordinator{
		chDB:               chDB,
		parser:             parser,
		tonClient:          tonClient,
		coordinatorAddress: coordinatorAddress,
	}
}

func (fetcher *FetcherEventsContractCoordinator) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.DefaultLogStartWork("FetcherEventsContractCoordinator: starting...")
	defer logger.Log.Info().Msg("FetcherEventsContractCoordinator: stopped")

	rawEventChan := make(chan *ton.RawEvent, 64)
	{
		fetcherEventCollector := ton.NewRawEventCollector(fetcher.tonClient, fetcher.coordinatorAddress, rawEventChan)

		wg.Add(1)
		go func() {
			defer wg.Done()
			fetcherEventCollector.Work(ctx, 10*time.Second) // TODO: move to config
		}()
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherEventsContractCoordinator received shutdown signal...")
			return
		case rawEvent, ok := <-rawEventChan:
			if !ok {
				logger.Log.Info().Msg("FetcherEventsContractCoordinator: channel closed")
				return
			}
			fetcher.Fetch(rawEvent)
			watchdog.Global().Heartbeat("FetcherEventsContractCoordinator")
		}
	}
}

func (fetcher *FetcherEventsContractCoordinator) Fetch(rawEvent *ton.RawEvent) {
	event, err := fetcher.parser.Parse(rawEvent)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherEventsContractCoordinator").
			Msg("failed to parse event")

		return
	}

	fetcher.chDB <- event
}
