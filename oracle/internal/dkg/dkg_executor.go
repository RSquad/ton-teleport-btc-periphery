package dkg

import (
	"context"
	"math"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type Executor struct {
	inChan chan *coordinatorcontract.DKG
	until  time.Time
}

func NewExecutor(inChan chan *coordinatorcontract.DKG) *Executor {
	return &Executor{
		inChan: inChan,
		until:  time.Unix(0, 0),
	}
}

func (e *Executor) Work(ctx context.Context) (err error) {
	logger.DefaultLogStartWork("DKGExecutor")
	defer logger.DefaultLogFinishWork("DKGExecutor", err)
	for {
		dkg, ok := <-e.inChan
		if !ok {
			return nil
		}
		e.Execute(dkg)
	}
}

func (e *Executor) Execute(dkg *coordinatorcontract.DKG) {
	e.logStartExecuting(dkg)
	defer e.logFinishExecuting(dkg)

	if dkg.Status == coordinatorcontract.DKGStatusFinished {
		e.logDKGFinished(dkg)
		return
	}

	if dkg.Until.After(e.until) {
		e.until = dkg.Until
		e.logNewDKGStarted(dkg)
	}
}

func (e *Executor) executeR1(dkg *coordinatorcontract.DKG, validatorIdx uint16, identifier []byte) {
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status == coordinatorcontract.DKGStatusPart1Finished {
		return
	}

	packages := dkg.R1.GetPkgs()
	if packages.Get(string(identifier)) != nil {
		return
	}

	// TODO: store local part1Result while it no sended to coordinator
	part1Result, _, err := frost.DkgPart1(identifier, uint16(math.Floor(float64(dkg.MaxSigners)*2/3)), uint16(dkg.MaxSigners))
	if err != nil {
		return
	}

	// TODO: add imports for correct call SendRound1 method
	coordinatorcontract.SendRound1(
		coordinatorcontract.DefaultDGKTTL,
		validatorIdx,
		identifier,
		part1Result,
	)

	return
}
