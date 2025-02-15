package ton

import (
	"context"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"golang.org/x/sync/errgroup"
)

type RawEventCollector struct {
	tonClient *tonclient.TonClient
	addr      *address.Address
	outChan   chan<- *RawEvent
}

func NewRawEventCollector(
	tonClient *tonclient.TonClient,
	addr *address.Address,
	outChan chan<- *RawEvent,
) *RawEventCollector {
	return &RawEventCollector{
		tonClient: tonClient,
		addr:      addr,
		outChan:   outChan,
	}
}

func (ec *RawEventCollector) Work(ctx context.Context) (err error) {
	ec.logStartWork()
	defer ec.logFinishWork(err)

	txChan := make(chan *tlb.Transaction)
	txCollector, err := tonclient.NewTxCollector(ec.tonClient, ec.addr, txChan)
	if err != nil {
		return err
	}
	eventFilter := NewRawEventFilter(txChan, ec.outChan)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return txCollector.Work(ctx)
	})

	g.Go(func() error {
		return eventFilter.Work(ctx)
	})

	if werr := g.Wait(); werr != nil {
		return werr
	}

	return nil
}
