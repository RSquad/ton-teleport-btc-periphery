package tonclient

import (
	"context"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"golang.org/x/sync/errgroup"
)

type TxCollector struct {
	tonClient    *TonClient
	addr         *address.Address
	txFetcher    *TxFetcher
	txSubscriber *TxSubscriber
}

func newTxCollector(
	tonClient *TonClient,
	addr *address.Address,
	outChan chan<- *tlb.Transaction,
	serverTimeout time.Duration,
	lastTxLT uint64,
	lastTxHash []byte,
) *TxCollector {
	txFetcher := NewTxFetcher(tonClient, addr, lastTxLT, lastTxHash, 16, outChan, serverTimeout)
	txSubscriber := NewTxSubscriber(tonClient, addr, lastTxLT, outChan)

	return &TxCollector{
		tonClient:    tonClient,
		addr:         addr,
		txFetcher:    txFetcher,
		txSubscriber: txSubscriber,
	}
}

func NewTxCollector(
	tonClient *TonClient,
	addr *address.Address,
	outChan chan<- *tlb.Transaction,
	serverTimeout time.Duration,
) (*TxCollector, error) {
	acc, err := tonClient.FetchAcc(addr, nil)
	if err != nil {
		return nil, err
	}

	return newTxCollector(tonClient, addr, outChan, serverTimeout, acc.LastTxLT, acc.LastTxHash), nil
}

func NewTxCollectorLT(
	tonClient *TonClient,
	addr *address.Address,
	outChan chan<- *tlb.Transaction,
	serverTimeout time.Duration,
	lastTxLT uint64,
) (*TxCollector, error) {
	acc, err := tonClient.FetchAcc(addr, nil)
	if err != nil {
		return nil, err
	}

	return newTxCollector(tonClient, addr, outChan, serverTimeout, lastTxLT, acc.LastTxHash), nil
}

func (tc *TxCollector) Work(ctx context.Context) (err error) {
	tc.logStartWork()
	defer func() {
		tc.logFinishWork(err)
	}()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return tc.txSubscriber.Work(ctx)
	})

	g.Go(func() error {
		return tc.txFetcher.Work(ctx)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
