package ton

import (
	"context"
	"time"

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

func (ef *RawEventFilter) Work(ctx context.Context) (err error) {
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
						ef.outChan <- &RawEvent{
							Addr:    tx.IO.In.Msg.DestAddr(),
							TxHash:  tx.Hash,
							TxLT:    tx.LT,
							TxUtime: time.Unix(int64(tx.Now), 0),
							Body:    msg.AsExternalOut().Body,
						}
					}
				}
			}
		}
	}
}
