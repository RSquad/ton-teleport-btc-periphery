package tonclient

import (
	"context"

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

func NewTxCollector(
	tonClient *TonClient,
	addr *address.Address,
	outChan chan<- *tlb.Transaction,
) (*TxCollector, error) {
	acc, err := tonClient.FetchAcc(addr, nil)
	if err != nil {
		return nil, err
	}

	txFetcher := NewTxFetcher(tonClient, addr, acc.LastTxLT, acc.LastTxHash, 64, outChan)
	txSubscriber := NewTxSubscriber(tonClient, addr, acc.LastTxLT, outChan)

	return &TxCollector{
		tonClient:    tonClient,
		addr:         addr,
		txFetcher:    txFetcher,
		txSubscriber: txSubscriber,
	}, nil
}

func (tc *TxCollector) Work(ctx context.Context) (err error) {
	tc.logStartWork()
	defer tc.logFinishWork(err)

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
