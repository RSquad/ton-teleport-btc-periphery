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
	inChan              chan *coordinatorcontract.DKG
	until               time.Time
	coordinatorContract *coordinatorcontract.CoordinatorContract
}

func NewExecutor(inChan chan *coordinatorcontract.DKG, coordinatorContract *coordinatorcontract.CoordinatorContract) *Executor {
	return &Executor{
		inChan:              inChan,
		until:               time.Unix(0, 0),
		coordinatorContract: coordinatorContract,
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
	e.logDKGProcess(dkg, "Start executing R1")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status >= coordinatorcontract.DKGStatusPart1Finished {
		e.logDKGProcess(dkg, "R1 completed")
		return
	}

	packages := dkg.R1.GetPkgs()
	if packages.Get(string(identifier)) != nil {
		e.logDKGProcess(dkg, "R1 package already sent")
		return
	}

	// TODO: store local part1Result while it no sended to coordinator
	part1Result, r1Secret, err := frost.DkgPart1(identifier, uint16(math.Floor(float64(dkg.MaxSigners)*2/3)), uint16(dkg.MaxSigners))
	if err != nil {
		return
	}

	// TODO: store part1Result (package), r1Secret

	e.coordinatorContract.SendRound1(
		int64(coordinatorcontract.DefaultDGKTTL),
		validatorIdx,
		identifier,
		part1Result,
	)
}

func (e *Executor) executeR2(dkg *coordinatorcontract.DKG, validatorIdx uint16, identifier []byte) {
	e.logDKGProcess(dkg, "Start executing R2")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status >= coordinatorcontract.DKGStatusPart2Finished {
		return
	}

	// part2Result, r2Secret, err := frost.DkgPart2(r1Secret, r1Pkgs)

	// if err != nil {
	// 	return
	// }

	// TODO: store part2Result (packages), r2Secret

	// for ident := range part2Result {
	// 	e.coordinatorContract.SendRound2(
	// 		int64(coordinatorcontract.DefaultDGKTTL),
	// 		validatorIdx,
	// 		identifier,
	// 		ident,
	// 		part2Result[ident],
	// 	)
	// }
}

func (e *Executor) executeR3(dkg *coordinatorcontract.DKG, validatorIdx uint16, identifier []byte) {
	e.logDKGProcess(dkg, "Start executing R3")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status == coordinatorcontract.DKGStatusPart1Finished {
		return
	}

	packages := dkg.R1.GetPkgs()
	if packages.Get(string(identifier)) != nil {
		return
	}

	// keyPackage, publicKeyPackage, err := frost.DkgPart3(r2Secret, r1Packages, r2Packages)
	// if err != nil {
	// 	return
	// }

	// // TODO: store keyPackage publicKeyPackage

	// e.coordinatorContract.SendPubkeyPackage(
	// 	int64(coordinatorcontract.DefaultDGKTTL),
	// 	validatorIdx,
	// 	keyPackage,
	// 	identifier,
	// 	publicKeyPackage,
	// )
}
