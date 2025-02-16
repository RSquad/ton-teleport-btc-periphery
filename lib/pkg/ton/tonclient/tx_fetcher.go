package tonclient

import (
	"context"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxFetcher struct {
	tonClient *TonClient
	addr      *address.Address
	lt        uint64
	hash      []byte
	limit     uint32
	outChan   chan<- *tlb.Transaction
}

func NewTxFetcher(
	tonClient *TonClient,
	addr *address.Address,
	lt uint64,
	hash []byte,
	limit uint32,
	outChan chan<- *tlb.Transaction,
) *TxFetcher {
	if limit == 0 {
		limit = 64
	}
	return &TxFetcher{
		tonClient: tonClient,
		addr:      addr,
		lt:        lt,
		hash:      hash,
		limit:     limit,
		outChan:   outChan,
	}
}

func (tf *TxFetcher) Work(ctx context.Context) (err error) {
	tf.logStartWork()

	count := 0
	start := time.Now()

	defer func() {
		tf.logFinishWork(count, time.Since(start), err)
	}()

	for {
		innerCtx, cancelInnerCtx := context.WithCancel(ctx)

		select {
		case <-ctx.Done():
			cancelInnerCtx()
			return ctx.Err()
		default:
		}

		txs, ferr := tf.Fetch(innerCtx)
		if ferr != nil {
			cancelInnerCtx()
			tf.logFetchError(ferr)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(txs) == 0 {
			cancelInnerCtx()
			return nil
		}

		tf.logTxsFetched(len(txs), count+len(txs))

		for _, tx := range txs {
			tf.outChan <- tx
			count++

			if tx.PrevTxLT < tf.lt {
				tf.lt = tx.PrevTxLT
				tf.hash = tx.PrevTxHash
			}
			if tf.lt == 0 {
				cancelInnerCtx()
				return nil
			}
		}
	}
}

func (tf *TxFetcher) Fetch(ctx context.Context) ([]*tlb.Transaction, error) {
	txs, err := tf.tonClient.API.ListTransactions(ctx, tf.addr, tf.limit, tf.lt, tf.hash)
	if err != nil {
		return nil, err
	}
	return txs, nil
}
