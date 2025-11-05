package fetchers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
)

type FetcherEventsContractCoordinator struct {
	inChan              chan *ton.RawEvent
	chDB                chan PayloadDB
	coordinatorContract coordinator.Coordinator
	period              int64 // Fetch period (in seconds)
	parser              ton.EventParserInterface
}

func NewFetcherEventsContractCoordinator(
	inChan chan *ton.RawEvent,
	chDB chan PayloadDB,
	coordinatorContract coordinator.Coordinator,
	period int64,
	parser ton.EventParserInterface,
) *FetcherEventsContractCoordinator {
	return &FetcherEventsContractCoordinator{
		inChan:              inChan,
		chDB:                chDB,
		coordinatorContract: coordinatorContract,
		period:              period,
		parser:              parser,
	}
}

func (fetcher *FetcherEventsContractCoordinator) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherEventsContractCoordinator: stopped")
	logger.DefaultLogStartWork("FetcherEventsContractCoordinator: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	watchdog.Global().Watch("FetcherEventsContractCoordinator", time.Duration(fetcher.period*2)*time.Second)
	defer watchdog.Global().Unwatch("FetcherEventsContractCoordinator")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherEventsContractCoordinator received shutdown signal...")
			return
		case <-ticker.C:
			rawEvent, ok := <-fetcher.inChan
			if !ok {
				logger.Log.Error().Msg("FetcherEventsContractCoordinator: wait")
				wg.Wait()
			}
			fetcher.Fetch(rawEvent)
			watchdog.Global().Heartbeat("FetcherEventsContractCoordinator")
		}
	}
}

func (fetcher *FetcherEventsContractCoordinator) Fetch(rawEvent *ton.RawEvent) {
	event, err := fetcher.parser.Parse(rawEvent)
	if err != nil {
		if strings.Contains(err.Error(), "unknown event type") {
			logger.Log.Error().Err(err).
				Str("component", "FetcherEventsContractCoordinator").
				Msg("failed to parse event")
			return
		}
		return
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherEventsContractCoordinator").
			Msg("failed to serialize event->json")

		return
	}

	fetcher.chDB <- PayloadDB{
		typeId:  PayloadTypeContractCoordinatorEvents,
		payload: string(jsonData),
	}
}
