package dkg

import (
	"log"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type Executor struct {
	until time.Time
}

func NewExecutor() *Executor {
	return &Executor{
		until: time.Unix(0, 0),
	}
}

func (h *Executor) Execute(dkg *coordinatorcontract.DKG) {
	if dkg.Status == coordinatorcontract.DKGStatusFinished {
		log.Printf("DKG finished")
		return
	}

	if dkg.Until.After(h.until) {
		h.until = dkg.Until
		log.Printf("New DKG started until %s", h.until)
	}
}
