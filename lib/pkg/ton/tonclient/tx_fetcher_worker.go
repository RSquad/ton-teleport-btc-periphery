package tonclient

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxFetcherWorker struct {
	fetcher  *TxFetcher
	addr     *address.Address
	lastLT   uint64
	lastHash []byte
	ctx      context.Context
	cancel   context.CancelFunc
	running  atomic.Bool
}

func NewTxFetcherWorker(fetcher *TxFetcher) *TxFetcherWorker {
	return &TxFetcherWorker{fetcher: fetcher}
}

func (tw *TxFetcherWorker) Run(addr *address.Address, lt uint64, hash []byte, outChan chan<- *tlb.Transaction) error {
	if !tw.running.CompareAndSwap(false, true) {
		return errors.New("worker already running")
	}
	defer tw.Stop()
	tw.initWorkerState(addr, lt, hash)
	log.Printf("start fetching txs (addr=%v, txhash=%x, txlt=%v)", utils.AddrToRawString(addr), hash, lt)

	txCount, startTime := 0, time.Now()

	for !tw.isCancelled() {
		txs, err := tw.fetch()
		if err != nil {
			tw.handleFetchError(err)
			continue
		}

		if len(txs) == 0 {
			tw.Stop()
		}

		txCount += tw.process(txs, outChan)
	}

	log.Printf("txs fetched (addr=%v, count=%v, duration=%v)", utils.AddrToRawString(tw.addr), txCount, time.Since(startTime))

	tw.clearWorkerState()

	return nil
}

func (tw *TxFetcherWorker) Stop() {
	if tw.cancel != nil {
		tw.cancel()
		log.Printf("tx fetch stopped (addr=%v)", utils.AddrToRawString(tw.addr))
	}
}

func (tw *TxFetcherWorker) initWorkerState(addr *address.Address, lt uint64, hash []byte) {
	ctx, cancel := context.WithCancel(context.Background())
	tw.ctx, tw.cancel, tw.addr, tw.lastLT, tw.lastHash = ctx, cancel, addr, lt, hash
}

func (tw *TxFetcherWorker) clearWorkerState() {
	*tw = TxFetcherWorker{fetcher: tw.fetcher}
}

func (tw *TxFetcherWorker) fetch() ([]*tlb.Transaction, error) {
	return tw.fetcher.Fetch(tw.ctx, tw.addr, tw.lastLT, tw.lastHash, 64)
}

func (tw *TxFetcherWorker) process(txs []*tlb.Transaction, outChan chan<- *tlb.Transaction) int {
	count := 0
	for _, tx := range txs {
		if tw.isCancelled() {
			return count
		}

		outChan <- tx
		count++

		if tw.shouldStopFetching(tx) {
			tw.Stop()
		}

		tw.updateLastTx(tx)
	}
	return count
}

func (tw *TxFetcherWorker) handleFetchError(err error) {
	log.Printf("error fetching txs (addr=%v, txhash=%x, txlt=%v): %v", utils.AddrToRawString(tw.addr), tw.lastHash, tw.lastLT, err)
	time.Sleep(1 * time.Second)
}

func (tw *TxFetcherWorker) updateLastTx(tx *tlb.Transaction) {
	if tx.PrevTxLT < tw.lastLT {
		tw.lastLT = tx.PrevTxLT
		tw.lastHash = tx.PrevTxHash
	}
}

func (tw *TxFetcherWorker) shouldStopFetching(tx *tlb.Transaction) bool {
	return tx.PrevTxLT == 0
}

func (tw *TxFetcherWorker) isCancelled() bool {
	select {
	case <-tw.ctx.Done():
		return true
	default:
		return false
	}
}
