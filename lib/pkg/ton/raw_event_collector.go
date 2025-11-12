package ton

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
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

func (ec *RawEventCollector) Work(ctx context.Context, serverTimeout time.Duration) (err error) {
	ec.logStartWork()
	defer func() {
		ec.logFinishWork(err)
	}()

	txChan := make(chan *tlb.Transaction, 128)
	defer close(txChan)
	txCollector, err := tonclient.NewTxCollector(ec.tonClient, ec.addr, txChan, serverTimeout)
	if err != nil {
		return err
	}
	eventFilter := NewRawEventFilter(txChan, ec.outChan)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := txCollector.Work(ctx)
		if err != nil {
			logger.Log.Debug().
				Str("component", "RawEventCollector").
				Err(err).
				Msg("txCollector finished")
		}
		return err
	})

	g.Go(func() error {
		err := eventFilter.Work(ctx)
		if err != nil {
			logger.Log.Debug().
				Str("component", "RawEventCollector").
				Err(err).
				Msg("eventFilter finished")
		}
		return err
	})

	if werr := g.Wait(); werr != nil {
		logger.Log.Debug().
			Str("component", "RawEventCollector").
			Err(werr).
			Msg("errgroup.Wait returned error")
		return werr
	}

	return nil
}
