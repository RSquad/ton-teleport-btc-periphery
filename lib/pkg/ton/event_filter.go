package ton

import (
	"log"
	"time"

	"github.com/xssnick/tonutils-go/tlb"
)

type EventFilter struct{}

func NewEventFilter() *EventFilter {
	return &EventFilter{}
}

func (tf *EventFilter) Run(txChan <-chan *tlb.Transaction, rawEventsChan chan<- *RawEvent) {
	for tx := range txChan {
		if tx.IO.Out != nil {
			outMsgs, err := tx.IO.Out.ToSlice()
			if err != nil {
				log.Printf("error extracting logs from tx (hash=%x lt=%v): %v", tx.Hash, tx.LT, err)
				continue
			}
			for _, msg := range outMsgs {
				if msg.MsgType == tlb.MsgTypeExternalOut {
					rawEventsChan <- &RawEvent{
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
