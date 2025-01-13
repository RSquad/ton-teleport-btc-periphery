package tonclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxFetcher struct {
	tonClient *TonClient
	addr      *address.Address
	lastLT    uint64
	lastHash  []byte
	ctx       context.Context
	cancelCtx context.CancelFunc
	mu        sync.Mutex
}

func NewTxFetcher(tonClient *TonClient, addr *address.Address) *TxFetcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &TxFetcher{
		tonClient: tonClient,
		addr:      addr,
		ctx:       ctx,
		cancelCtx: cancel,
	}
}

func (tf *TxFetcher) Run(txChan chan<- *tlb.Transaction, lt uint64, hash []byte) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	log.Printf("fetching txs (addr=%v, txhash=%x, txlt=%v)", utils.AddrToRawString(tf.addr), hash, lt)

	tf.lastLT = lt
	tf.lastHash = hash

	txCount := 0
	startTime := time.Now()

	for {
		if tf.isCtxCancelled() {
			break
		}

		txs, err := tf.fetchTxs()
		if err != nil {
			time.Sleep(3 * time.Second)
			log.Print(err)
			continue
		}

		tf.publishTxs(txs, txChan)
		txCount += len(txs)
	}

	log.Printf(
		"txs fetched (addr=%v, txcount=%v, duration=%v)", utils.AddrToRawString(tf.addr), txCount, time.Since(startTime))
}

func (tf *TxFetcher) Kill() {
	if tf.cancelCtx != nil {
		tf.cancelCtx()
	}
}

func (tf *TxFetcher) fetchTxs() ([]*tlb.Transaction, error) {
	txs, err := tf.tonClient.API.WithRetry(3).
		ListTransactions(tf.ctx, tf.addr, 64, tf.lastLT, tf.lastHash)
	if err != nil {
		return nil, fmt.Errorf(
			"error fetching txs (addr=%v, txhash=%x, txlt=%v): %w",
			utils.AddrToRawString(tf.addr),
			tf.lastHash,
			tf.lastLT,
			err,
		)
	}
	return txs, nil
}

func (tf *TxFetcher) publishTxs(
	txs []*tlb.Transaction,
	txChan chan<- *tlb.Transaction,
) {
	for _, tx := range txs {
		if tf.isCtxCancelled() {
			return
		}
		txChan <- tx

		if tx.PrevTxLT == 0 {
			tf.Kill()
			return
		}
		if tx.PrevTxLT < tf.lastLT {
			tf.lastLT = tx.PrevTxLT
			tf.lastHash = tx.PrevTxHash
		}
	}
}

func (tf *TxFetcher) isCtxCancelled() bool {
	select {
	case <-tf.ctx.Done():
		return true
	default:
		return false
	}
}
