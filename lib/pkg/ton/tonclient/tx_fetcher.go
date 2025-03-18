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
		time.Sleep(150 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		shortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		txs, ferr := tf.Fetch(shortCtx)
		cancel()

		if ferr != nil {
			tf.logFetchError(ferr)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if len(txs) == 0 {
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
		}

		if tf.lt == 0 {
			return nil
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
