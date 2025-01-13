package ton

import (
	"context"
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
	block, err := ec.tonClient.API.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil
	}

	contractAcc, err := ec.tonClient.API.GetAccount(context.Background(), block, ec.contract.GetAddr())
	if err != nil {
		return nil
	}

	txFetcher := tonclient.NewTxFetcher(ec.tonClient, ec.contract.GetAddr())
	txSubscriber := tonclient.NewTxSubscriber(ec.tonClient, ec.contract.GetAddr(), contractAcc.LastTxLT)
	eventFilter := NewEventFilter()

	txChan := make(chan *tlb.Transaction)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		txFetcher.Run(txChan, contractAcc.LastTxLT, contractAcc.LastTxHash)
	}()
	go func() {
		defer wg.Done()
		txSubscriber.Run(txChan)
	}()
	go func() {
		defer wg.Done()
		eventFilter.Run(txChan, rawEventChan)
	}()

	wg.Wait()

	return nil
}
