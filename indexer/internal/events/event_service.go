package events

import (
	"context"

	"golang.org/x/sync/errgroup"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type EventService struct {
	tonClient           *tonclient.TonClient
	repo                *ent.Client
	teleportContract    *teleportcontract.TeleportContract
	coordinatorContract *coordinatorcontract.CoordinatorContract
}

func NewEventService(
	tonClient *tonclient.TonClient,
	repo *ent.Client,
	teleportContract *teleportcontract.TeleportContract,
	coordinatorContract *coordinatorcontract.CoordinatorContract,
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
	defer es.logFinishWork(err)

	teleportContractStorage, err := es.teleportContract.GetStorage(nil)
	if err != nil {
		return err
	}

	rawEventChan := make(chan *ton.RawEvent, 64)
	teleportContractRawEventCollector := es.createRawEventCollector(es.teleportContract.GetAddr(), rawEventChan)
	coordinatorContractRawEventCollector := es.createRawEventCollector(es.coordinatorContract.GetAddr(), rawEventChan)

	tonTxWriter := es.createTonTxWriter(ctx)
	pegoutWriter := es.createPegoutWriter(ctx, teleportContractStorage.PegoutContractCode)
	eventWriter := es.createEventWriter(ctx, pegoutWriter)

	dispatcher := es.createEventDispatcher(rawEventChan, tonTxWriter, eventWriter)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return teleportContractRawEventCollector.Work(ctx)
	})

	g.Go(func() error {
		return coordinatorContractRawEventCollector.Work(ctx)
	})

	g.Go(func() error {
		return dispatcher.Work(ctx)
	})

	if werr := g.Wait(); werr != nil {
		return werr
	}

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
			es.coordinatorContract.GetAddr().String(): coordinatorcontract.NewEventParser(),
		},
	)
}
