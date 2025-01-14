package tonclient

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxSubscriberWorker struct {
	tonClient *TonClient
	addr      *address.Address
	lt        uint64
	ctx       context.Context
	cancel    context.CancelFunc
	running   atomic.Bool
}

func NewTxSubscriberWorker(tonClient *TonClient) *TxSubscriberWorker {
	return &TxSubscriberWorker{tonClient: tonClient}
}

func (ts *TxSubscriberWorker) Run(addr *address.Address, lt uint64, outChan chan<- *tlb.Transaction) error {
	if !ts.running.CompareAndSwap(false, true) {
		return errors.New("worker already running")
	}
	defer ts.Stop()

	ts.initWorkerState(addr, lt)
	log.Printf("subscribed to txs (addr=%v, lt=%v)", utils.AddrToRawString(ts.addr), ts.lt)

	subChan := make(chan *tlb.Transaction)
	go ts.subscribe(subChan)

	ts.process(subChan, outChan)

	log.Printf("stopped subscription to txs (addr=%v)", utils.AddrToRawString(ts.addr))

	return nil
}

func (ts *TxSubscriberWorker) Stop() {
	if ts.cancel != nil {
		ts.cancel()
	}
	ts.clearWorkerState()
	log.Printf("tx subscription stopped (addr=%v)", utils.AddrToRawString(ts.addr))
}

func (ts *TxSubscriberWorker) initWorkerState(addr *address.Address, lt uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	ts.ctx, ts.cancel, ts.addr, ts.lt = ctx, cancel, addr, lt
}

func (ts *TxSubscriberWorker) clearWorkerState() {
	*ts = TxSubscriberWorker{tonClient: ts.tonClient}
	ts.running.Store(false)
}

func (ts *TxSubscriberWorker) subscribe(subChan chan<- *tlb.Transaction) {
	ts.tonClient.API.SubscribeOnTransactions(ts.ctx, ts.addr, ts.lt, subChan)
}

func (ts *TxSubscriberWorker) process(subChan <-chan *tlb.Transaction, outChan chan<- *tlb.Transaction) {
	for {
		select {
		case <-ts.ctx.Done():
			return
		case tx, ok := <-subChan:
			if !ok {
				return
			}
			ts.publish(tx, outChan)
		}
	}
}

func (ts *TxSubscriberWorker) publish(tx *tlb.Transaction, outChan chan<- *tlb.Transaction) {
	outChan <- tx
}
