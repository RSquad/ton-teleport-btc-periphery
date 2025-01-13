package events

import (
	"context"
	"log"
	"sync"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
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
		tonClient, repo, teleportContract, coordinatorContract,
	}
}

func (es *EventService) Run() {
	teleportContractRawEventCollector := ton.NewRawEventCollector(
		es.tonClient,
		es.teleportContract,
	)
	coordinatorContractRawEventCollector := ton.NewRawEventCollector(
		es.tonClient,
		es.coordinatorContract,
	)

	teleportContractEventParserExecutor := ton.NewEventParserExecutor(teleportcontract.NewEventParser())
	coordinatorContractEventParserExecutor := ton.NewEventParserExecutor(coordinatorcontract.NewEventParser())

	storage, err := es.teleportContract.GetStorage(nil)
	if err != nil {
		log.Printf("failed to get teleport contract storage: %v", err)
	}
	pegoutWriter := pegout.NewPegoutWriter(context.Background(), es.repo, es.teleportContract, storage.PegoutContractCode)

	eventWriter := NewEventWriter(context.Background(), es.repo, pegoutWriter)
	eventWriterExecutor := NewEventWriterExecutor(eventWriter)

	teleportContractRawEventChan := make(chan *ton.RawEvent)
	coordinatorContractRawEventChan := make(chan *ton.RawEvent)
	eventChan := make(chan ton.EventInterface)

	var wg sync.WaitGroup

	wg.Add(5)
	go func() {
		defer wg.Done()
		teleportContractRawEventCollector.Run(teleportContractRawEventChan)
	}()
	go func() {
		defer wg.Done()
		coordinatorContractRawEventCollector.Run(coordinatorContractRawEventChan)
	}()
	go func() {
		defer wg.Done()
		teleportContractEventParserExecutor.Run(teleportContractRawEventChan, eventChan)
	}()
	go func() {
		defer wg.Done()
		coordinatorContractEventParserExecutor.Run(coordinatorContractRawEventChan, eventChan)
	}()
	go func() {
		defer wg.Done()
		eventWriterExecutor.Run(eventChan)
	}()

	wg.Wait()
}
