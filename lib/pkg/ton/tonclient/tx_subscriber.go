package tonclient

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxSubscriber struct {
	tonClient *TonClient
	addr      *address.Address
	lt        uint64
	outChan   chan<- *tlb.Transaction
}

func NewTxSubscriber(
	tonClient *TonClient,
	addr *address.Address,
	lt uint64,
	outChan chan<- *tlb.Transaction,
) *TxSubscriber {
	return &TxSubscriber{
		tonClient: tonClient,
		addr:      addr,
		lt:        lt,
		outChan:   outChan,
	}
}

func (ts *TxSubscriber) Work(ctx context.Context) (err error) {
	ts.logStartWork()
	defer func() {
		ts.logFinishWork(err)
	}()

	subChan := make(chan *tlb.Transaction)
	go ts.tonClient.API.SubscribeOnTransactions(ctx, ts.addr, ts.lt, subChan)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tx, ok := <-subChan:
			if !ok {
				return nil
			}
			ts.logTxReceived(tx)
			if err := ts.writeToChannel(ctx, tx); err != nil {
				return err
			}
		}
	}
}

func (ts *TxSubscriber) writeToChannel(ctx context.Context, tx *tlb.Transaction) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ts.outChan <- tx:
			return nil
		case <-time.After(5 * time.Second):
			logger.Log.Debug().Str("component", "TxSubscriber").Msg("channel is full, retrying...")
		}
	}
}
