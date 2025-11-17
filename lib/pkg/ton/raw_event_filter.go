package ton

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/xssnick/tonutils-go/tlb"
)

type RawEventFilter struct {
	inChan  <-chan *tlb.Transaction
	outChan chan<- *RawEvent
}

func NewRawEventFilter(
	inChan <-chan *tlb.Transaction,
	outChan chan<- *RawEvent,
) *RawEventFilter {
	return &RawEventFilter{
		inChan:  inChan,
		outChan: outChan,
	}
}

func (ef *RawEventFilter) writeToChannel(ctx context.Context, event *RawEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ef.outChan <- event:
			return nil
		case <-time.After(5 * time.Second):
			logger.Log.Debug().Str("component", "RawEventFilter").Msg("channel is full, retrying...")
		}
	}
}

func (ef *RawEventFilter) Work(ctx context.Context) (err error) {
	ef.logStartWork()
	defer func() {
		ef.logFinishWork(err)
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tx, ok := <-ef.inChan:
			if !ok {
				return nil
			}
			if tx.IO.Out != nil {
				outMsgs, err := tx.IO.Out.ToSlice()
				if err != nil {
					ef.logFilterError(tx, err)
					continue
				}
				for _, msg := range outMsgs {
					if msg.MsgType == tlb.MsgTypeExternalOut {
						event := &RawEvent{
							Addr:    tx.IO.In.Msg.DestAddr(),
							TxHash:  tx.Hash,
							TxLT:    tx.LT,
							TxUtime: time.Unix(int64(tx.Now), 0),
							Body:    msg.AsExternalOut().Body,
						}
						if err := ef.writeToChannel(ctx, event); err != nil {
							return err
						}
					}
				}
			}
		}
	}
}
