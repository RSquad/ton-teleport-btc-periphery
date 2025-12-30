package events

import (
	"context"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type EventService struct {
	tonClient           *tonclient.TonClient
	repo                *ent.Client
	teleportContract    *teleportcontract.TeleportContract
	coordinatorContract coordinator.Coordinator
}

func NewEventService(
	tonClient *tonclient.TonClient,
	repo *ent.Client,
	teleportContract *teleportcontract.TeleportContract,
	coordinatorContract coordinator.Coordinator,
) *EventService {
	return &EventService{
		tonClient:           tonClient,
		repo:                repo,
		teleportContract:    teleportContract,
		coordinatorContract: coordinatorContract,
	}
}

func (es *EventService) Work(ctx context.Context) (err error) {
	es.logStartWork()
	defer func() {
		es.logFinishWork(err)
	}()

	teleportContractStorage, err := es.teleportContract.GetStorage(nil)
	if err != nil {
		return ctx.Err()
	}

	rawEventChan := make(chan *ton.RawEvent, 64)
	defer close(rawEventChan)
	teleportContractRawEventCollector := es.createRawEventCollector(es.teleportContract.GetAddr(), rawEventChan)
	coordinatorContractRawEventCollector := es.createRawEventCollector(es.coordinatorContract.GetAddr(), rawEventChan)

	tonTxWriter := es.createTonTxWriter(ctx)
	pegoutWriter := es.createPegoutWriter(ctx, teleportContractStorage.PegoutContractCode)
	eventWriter := es.createEventWriter(ctx, pegoutWriter)

	dispatcher := es.createEventDispatcher(rawEventChan, tonTxWriter, eventWriter)

	logger.Log.Info().
		Str("component", "EventService").
		Msg("run event workers")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		teleportContractRawEventCollector.Work(ctx, config.Get().ServerTimeout)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		coordinatorContractRawEventCollector.Work(ctx, config.Get().ServerTimeout)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatcher.Work(ctx)
	}()

	wg.Wait()

	logger.Log.Info().
		Str("component", "EventService").
		Msg("event workers finished")
	return nil
}

func (es *EventService) createRawEventCollector(
	addr *address.Address,
	rawEventChan chan *ton.RawEvent,
) *ton.RawEventCollector {
	return ton.NewRawEventCollector(es.tonClient, addr, rawEventChan)
}

func (es *EventService) createTonTxWriter(
	ctx context.Context,
) *tontx.TonTxWriter {
	return tontx.NewTonTxWriter(ctx, es.repo)
}

func (es *EventService) createPegoutWriter(
	ctx context.Context,
	pegoutContractCode *cell.Cell,
) *pegout.PegoutWriter {
	return pegout.NewPegoutWriter(ctx, es.repo, es.teleportContract.GetAddr(), pegoutContractCode)
}

func (es *EventService) createEventWriter(
	ctx context.Context,
	pegoutWriter *pegout.PegoutWriter,
) *EventWriter {
	return NewEventWriter(ctx, es.repo, pegoutWriter)
}

func (es *EventService) createEventDispatcher(
	rawEventChan chan *ton.RawEvent,
	tonTxWriter *tontx.TonTxWriter,
	eventWriter *EventWriter,
) *EventDispatcher {
	return NewEventDispatcher(
		rawEventChan,
		tonTxWriter,
		eventWriter,
		map[string]ton.EventParserInterface{
			es.teleportContract.GetAddr().String():    teleportcontract.NewEventParser(),
			es.coordinatorContract.GetAddr().String(): coordinator.NewEventParser(),
		},
	)
}
