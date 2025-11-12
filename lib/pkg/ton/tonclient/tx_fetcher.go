package tonclient

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

var ErrNoMoreTransactions = errors.New("no more transactions")

type TxFetcher struct {
	tonClient     *TonClient
	addr          *address.Address
	lt            uint64
	hash          []byte
	limit         uint32
	outChan       chan<- *tlb.Transaction
	serverTimeout time.Duration
	interval      time.Duration
}

func NewTxFetcher(
	tonClient *TonClient,
	addr *address.Address,
	lt uint64,
	hash []byte,
	limit uint32,
	outChan chan<- *tlb.Transaction,
	serverTimeout time.Duration,
) *TxFetcher {
	if limit == 0 {
		// maximum limit in Testnet LiteServer
		limit = 16
	}
	return &TxFetcher{
		tonClient:     tonClient,
		addr:          addr,
		lt:            lt,
		hash:          hash,
		limit:         limit,
		outChan:       outChan,
		serverTimeout: serverTimeout,
		interval:      2 * time.Second,
	}
}

func (tf *TxFetcher) fetch(ctx context.Context) ([]*tlb.Transaction, error) {
	apiCtx, cancel := context.WithTimeout(ctx, tf.serverTimeout)
	txs, err := tf.tonClient.API.ListTransactions(apiCtx, tf.addr, tf.limit, tf.lt, tf.hash)
	cancel()

	if err != nil {
		if errors.Is(err, ton.ErrNoTransactionsWereFound) {
			return nil, ErrNoMoreTransactions
		}
		return nil, fmt.Errorf("lite server error: %v", err)
	}

	if len(txs) == 0 {
		return nil, ErrNoMoreTransactions
	}

	return txs, nil
}

func (tf *TxFetcher) Work(ctx context.Context) (err error) {
	tf.logStartWork()

	count, start := 0, time.Now()
	defer func() {
		tf.logFinishWork(count, time.Since(start), err)
	}()

	ticker := time.NewTicker(tf.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			logger.Log.Debug().Str("component", "TxFetcher").Msg("fetching transactions")
			txs, err := tf.fetch(ctx)
			if err != nil {
				if errors.Is(err, ErrNoMoreTransactions) {
					logger.Log.Info().Str("component", "TxFetcher").Msg("no more transactions found")
					return nil
				}
				tf.logFetchError(err)
				continue
			}

			slices.SortFunc(txs, func(a, b *tlb.Transaction) int {
				return int(b.LT - a.LT)
			})

			tf.logTxsFetched(txs, count)
			for _, tx := range txs {
				// Retry sending until successful or context is cancelled
				if err := tf.writeToChannel(ctx, tx); err != nil {
					return err
				}
				count++
			}
			tf.lt = txs[len(txs)-1].PrevTxLT
			tf.hash = txs[len(txs)-1].PrevTxHash
		}
	}
}

func (tf *TxFetcher) writeToChannel(ctx context.Context, tx *tlb.Transaction) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tf.outChan <- tx:
			return nil
		case <-time.After(5 * time.Second):
			logger.Log.Debug().Str("component", "TxFetcher").Msg("channel is full, retrying...")
		}
	}
}
