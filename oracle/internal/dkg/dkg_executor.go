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
	pkg              []byte
	publicKeyPackage []byte
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

func (e *Executor) executeR1(dkg *coordinatorcontract.DKG, validatorIdx uint16, localIdentifier []byte) {
	e.logDKGProcess(dkg, "Start executing R1")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status >= coordinatorcontract.DKGStatusPart1Finished {
		e.logDKGProcess(dkg, "R1 completed")
		return
	}

	packages := dkg.GetR1Packages()
	if packages[string(localIdentifier)] != nil {
		e.logDKGProcess(dkg, "R1 package already sent")
		return
	}

	if e.frostState.r1 == nil {
		r1Package, r1Secret, err := frost.DkgPart1(
			localIdentifier,
			uint16(math.Floor(float64(dkg.MaxSigners)*2/3)),
			uint16(dkg.MaxSigners),
		)
		if err != nil {
			return
		}
		e.frostState.r1 = &r1Step{
			pkg:    r1Package,
			secret: r1Secret,
		}
	}

	e.coordinatorContract.SendRound1(
		int64(coordinatorcontract.DefaultDGKTTL),
		validatorIdx,
		localIdentifier,
		e.frostState.r1.pkg,
	)
}

func (e *Executor) executeR2(dkg *coordinatorcontract.DKG, validatorIdx uint16, localIdentifier []byte) {
	e.logDKGProcess(dkg, "Start executing R2")
	if dkg.Status == coordinatorcontract.DKGStatusFinished ||
		dkg.Status >= coordinatorcontract.DKGStatusPart2Finished {
		e.logDKGProcess(dkg, "R2 completed")
		return
	}

	if e.frostState.r2 == nil {
		r1Pkgs := map[frost.Identifier]frost.Package{}

		for identifier, pkg := range dkg.GetR1Packages() {
			if identifier == string(localIdentifier) {
				continue
			}

			r1Pkgs[frost.Identifier([]byte(identifier))] = frost.NewPackage(pkg)
		}

		r2Packages, r2Secret, err := frost.DkgPart2(e.frostState.r1.secret, r1Pkgs)
		if err != nil {
			return
		}
		e.frostState.r2 = &r2Step{
			pkgs:   r2Packages,
			secret: r2Secret,
		}
	}

	for identifierTo := range e.frostState.r2.pkgs {
		e.coordinatorContract.SendRound2(
			int64(coordinatorcontract.DefaultDGKTTL),
			validatorIdx,
			localIdentifier,
			identifierTo.ToBytes(),
			e.frostState.r2.pkgs[identifierTo].ToBytes(),
		)
	}
}

func (e *Executor) executeR3(dkg *coordinatorcontract.DKG, validatorIdx uint16, localIdentifier []byte) {
	e.logDKGProcess(dkg, "Start executing R3")
	if dkg.Status == coordinatorcontract.DKGStatusFinished {
		e.logDKGProcess(dkg, "R3 completed")
		return
	}

	if dkg.Status < coordinatorcontract.DKGStatusPart2Finished {
		e.logDKGProcess(dkg, "R2 not yet completed, waiting for more packages.")
		return
	}

	if e.frostState.r3 == nil {

		r1Packages := map[frost.Identifier]frost.Package{}

		for identifier, pkg := range dkg.GetR1Packages() {
			if identifier == string(localIdentifier) {
				continue
			}

			r1Packages[frost.Identifier([]byte(identifier))] = frost.NewPackage(pkg)
		}

		r2Packages := map[frost.Identifier]frost.Package{}

		for toIdentifier, pkg := range dkg.GetR2Packages(localIdentifier) {
			if toIdentifier == string(localIdentifier) {
				continue
			}
			r2Packages[frost.Identifier([]byte(toIdentifier))] = frost.NewPackage(pkg)
		}

		keyPackage, publicKeyPackage, err := frost.DkgPart3(e.frostState.r2.secret, r1Packages, r2Packages)
		if err != nil {
			return
		}
		e.frostState.r3 = &r3Step{
			pkg:              keyPackage,
			publicKeyPackage: publicKeyPackage,
		}
	}

	e.coordinatorContract.SendPubkeyPackage(
		int64(coordinatorcontract.DefaultDGKTTL),
		validatorIdx,
		e.frostState.r3.pkg,
		localIdentifier,
		e.frostState.r3.publicKeyPackage,
	)
}
