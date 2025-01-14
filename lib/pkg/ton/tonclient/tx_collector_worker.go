package tonclient

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxCollectorWorker struct {
	tonClient          *TonClient
	txSubscriberWorker *TxSubscriberWorker
	txFetcherWorker    *TxFetcherWorker
	ctx                context.Context
	cancel             context.CancelFunc
	running            atomic.Bool
}

func NewTxCollectorWorker(tonClient *TonClient) *TxCollectorWorker {
	txFetcher := NewTxFetcher(tonClient)
	txSubscriberWorker := NewTxSubscriberWorker(tonClient)
	txFetcherWorker := NewTxFetcherWorker(txFetcher)

	return &TxCollectorWorker{
		tonClient:          tonClient,
		txSubscriberWorker: txSubscriberWorker,
		txFetcherWorker:    txFetcherWorker,
	}
}

func (cw *TxCollectorWorker) Run(addr *address.Address, outChan chan<- *tlb.Transaction) error {
	if !cw.running.CompareAndSwap(false, true) {
		return errors.New("collector already running")
	}
	defer cw.Stop()

	cw.initWorkerState()

	acc, err := cw.fetchAcc(addr)
	if err != nil {
		return err
	}

	cw.startWorkers(addr, acc.LastTxLT, acc.LastTxHash, outChan)

	return nil
}

func (cw *TxCollectorWorker) Stop() {
	if cw.cancel != nil {
		cw.cancel()
	}
	cw.txFetcherWorker.Stop()
	cw.txSubscriberWorker.Stop()
	cw.clearWorkerState()
}

func (cw *TxCollectorWorker) initWorkerState() {
	ctx, cancel := context.WithCancel(context.Background())
	cw.ctx, cw.cancel = ctx, cancel
}

func (cw *TxCollectorWorker) clearWorkerState() {
	*cw = TxCollectorWorker{tonClient: cw.tonClient, txFetcherWorker: cw.txFetcherWorker, txSubscriberWorker: cw.txSubscriberWorker}
	cw.running.Store(false)
}

func (cw *TxCollectorWorker) fetchAcc(addr *address.Address) (*tlb.Account, error) {
	block, err := cw.tonClient.API.CurrentMasterchainInfo(cw.ctx)
	if err != nil {
		return nil, err
	}

	acc, err := cw.tonClient.API.GetAccount(cw.ctx, block, addr)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (cw *TxCollectorWorker) startWorkers(addr *address.Address, lt uint64, hash []byte, outChan chan<- *tlb.Transaction) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := cw.txFetcherWorker.Run(addr, lt, hash, outChan); err != nil {
			log.Printf("error running fetcher worker: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := cw.txSubscriberWorker.Run(addr, lt, outChan); err != nil {
			log.Printf("error running subscriber worker: %v", err)
		}
	}()

	wg.Wait()
}
