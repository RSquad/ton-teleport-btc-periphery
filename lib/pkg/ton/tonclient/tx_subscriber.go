package tonclient

import (
	"context"
	"log"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxSubscriber struct {
	tonClient *TonClient
	addr      *address.Address
	lt        uint64
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func NewTxSubscriber(
	tonClient *TonClient,
	addr *address.Address,
	lt uint64,
) *TxSubscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &TxSubscriber{
		tonClient: tonClient,
		addr:      addr,
		lt:        lt,
		ctx:       ctx,
		cancelCtx: cancel,
	}
}

func (ts *TxSubscriber) Run(txChan chan<- *tlb.Transaction) {
	log.Printf("subscribing to txs (addr=%v, lt=%v)", utils.AddrToRawString(ts.addr), ts.lt)
	_txChan := make(chan *tlb.Transaction)
	go ts.tonClient.API.SubscribeOnTransactions(ts.ctx, ts.addr, ts.lt, _txChan)

	go func() {
		for tx := range _txChan {
			txChan <- tx
		}
	}()
}

func (ts *TxSubscriber) Kill() {
	if ts.cancelCtx != nil {
		ts.cancelCtx()
	}
}
