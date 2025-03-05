package dkg

import (
	"context"
	"math"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type r1Step struct {
	pkg    []byte
	secret uintptr
}

type r2Step struct {
	pkgs   map[frost.Identifier]frost.Package
	secret uintptr
}

type r3Step struct {
	pkg    []byte
	secret uintptr
}

type ExecutorState struct {
	r1 *r1Step
	r2 *r2Step
	r3 *r3Step
}

type Executor struct {
	inChan              chan *coordinatorcontract.DKG
	until               time.Time
	coordinatorContract *coordinatorcontract.CoordinatorContract
	frostState          *ExecutorState
}

func NewExecutor(inChan chan *coordinatorcontract.DKG, coordinatorContract *coordinatorcontract.CoordinatorContract) *Executor {
	return &Executor{
		inChan:              inChan,
		until:               time.Unix(0, 0),
		coordinatorContract: coordinatorContract,
		frostState:          &ExecutorState{},
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

	if e.frostState.r1 == nil {
		part1Result, r1Secret, err := frost.DkgPart1(identifier, uint16(math.Floor(float64(dkg.MaxSigners)*2/3)), uint16(dkg.MaxSigners))
		if err != nil {
			return
		}
		e.frostState.r1 = &r1Step{
			pkg:    part1Result,
			secret: r1Secret,
		}
	}

	e.coordinatorContract.SendRound1(
		int64(coordinatorcontract.DefaultDGKTTL),
		validatorIdx,
		identifier,
		e.frostState.r1.pkg,
	)
}

func (e *Executor) executeR2(dkg *coordinatorcontract.DKG, validatorIdx uint16, identifier []byte) {
	e.logDKGProcess(dkg, "Start executing R2")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status >= coordinatorcontract.DKGStatusPart2Finished {
		e.logDKGProcess(dkg, "R2 completed")
		return
	}

	if e.frostState.r2 == nil {
		r1Pkgs := map[frost.Identifier]frost.Package{}

		for ident, pkg := range dkg.R1.GetPkgs().GetAll() {
			if ident == string(identifier) {
				continue
			}

			r1Pkgs[frost.Identifier([]byte(ident))] = frost.Package(pkg)
		}

		pkgs, r2Secret, err := frost.DkgPart2(e.frostState.r1.secret, r1Pkgs)
		if err != nil {
			return
		}
		e.frostState.r2 = &r2Step{
			pkgs:   pkgs,
			secret: r2Secret,
		}
	}

	for identifierTo := range e.frostState.r2.pkgs {
		e.coordinatorContract.SendRound2(
			int64(coordinatorcontract.DefaultDGKTTL),
			validatorIdx,
			identifier,
			[]byte(identifierTo),
			e.frostState.r2.pkgs[identifierTo].buf,
		)
	}
}

func (e *Executor) executeR3(dkg *coordinatorcontract.DKG, validatorIdx uint16, identifier []byte) {
	e.logDKGProcess(dkg, "Start executing R3")
	if dkg.Status == coordinatorcontract.DKGStatusFinished {
		e.logDKGProcess(dkg, "R3 completed")
		return
	}

	if e.frostState.r3 == nil {
		// TODO: get r1 and r2 packages from coordinator
		// keyPackage, publicKeyPackage, err := frost.DkgPart3(e.frostState.r2.secret, r1Packages, r2Packages)
		// if err != nil {
		// 	return
		// }
	}

	// e.coordinatorContract.SendPubkeyPackage(
	// 	int64(coordinatorcontract.DefaultDGKTTL),
	// 	validatorIdx,
	// 	keyPackage,
	// 	identifier,
	// 	publicKeyPackage,
	// )
}
