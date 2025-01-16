package ton

import (
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type RawEvent struct {
	Addr    *address.Address
	TxHash  []byte
	TxLT    uint64
	TxUtime time.Time
	Body    *cell.Cell
}

type EventInterface interface {
	GetEventID() uint32
	GetRaw() *RawEvent
}
