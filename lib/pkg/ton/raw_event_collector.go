package ton

import (
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/tlb"
)

type RawEventCollector struct {
	tonClient *tonclient.TonClient
	contract  ContractInterface
}

func NewRawEventCollector(
	tonClient *tonclient.TonClient,
	contract ContractInterface,
) *RawEventCollector {
	return &RawEventCollector{
		tonClient, contract,
	}
}

func (ec *RawEventCollector) Run(rawEventChan chan<- *RawEvent) error {
	txCollectorWorker := tonclient.NewTxCollectorWorker(ec.tonClient)
	eventFilter := NewEventFilter()

	txChan := make(chan *tlb.Transaction)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		txCollectorWorker.Run(ec.contract.GetAddr(), txChan)
	}()
	go func() {
		defer wg.Done()
		eventFilter.Run(txChan, rawEventChan)
	}()

	wg.Wait()

	return nil
}
