package dkg

import (
	"context"
	"math"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

type Secret struct {
	ptr uintptr
}

func NewSecret(ptr uintptr) Secret {
	return Secret{ptr: ptr}
}

type Round1Result struct {
	pkg    []byte
	secret Secret
}

type Round2Result struct {
	pkgs   map[frost.Identifier]frost.Package
	secret Secret
}

type Round3Result struct {
	pkg              []byte
	publicKeyPackage []byte
}

type ExecutionArtifacts struct {
	r1 *Round1Result
	r2 *Round2Result
	r3 *Round3Result
}

type Executor struct {
	inChan              chan *coordinatorcontract.DKG
	until               time.Time
	coordinatorContract *coordinatorcontract.CoordinatorContract
	artifacts           ExecutionArtifacts
}

func NewExecutor(inChan chan *coordinatorcontract.DKG, coordinatorContract *coordinatorcontract.CoordinatorContract) *Executor {
	return &Executor{
		inChan:              inChan,
		until:               time.Unix(0, 0),
		coordinatorContract: coordinatorContract,
		artifacts:           ExecutionArtifacts{},
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
	if dkg.Round1Completed() {
		e.logDKGProcess(dkg, "R1 completed")
		return
	}

	packages := dkg.GetR1Packages()
	if packages[string(localIdentifier)] != nil {
		e.logDKGProcess(dkg, "R1 package already sent")
		return
	}

	if e.artifacts.r1 == nil {
		r1Package, r1SecretPtr, err := frost.DkgPart1(
			localIdentifier,
			uint16(math.Floor(float64(dkg.MaxSigners)*2/3)),
			uint16(dkg.MaxSigners),
		)
		if err != nil {
			e.logDKGPart1Failed(dkg, err)
			return
		}
		e.artifacts.r1 = &Round1Result{
			pkg:    r1Package,
			secret: NewSecret(r1SecretPtr),
		}
	}

	e.coordinatorContract.SendRound1(
		validatorIdx,
		localIdentifier,
		e.artifacts.r1.pkg,
	)
}

func (e *Executor) executeR2(dkg *coordinatorcontract.DKG, validatorIdx uint16, localIdentifier []byte) {
	e.logDKGProcess(dkg, "Start executing R2")
	if dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 completed")
		return
	}

	if e.artifacts.r2 == nil {
		r1Pkgs := map[frost.Identifier]frost.Package{}

		for identifier, pkg := range dkg.GetR1Packages() {
			if identifier == string(localIdentifier) {
				continue
			}

			r1Pkgs[frost.Identifier([]byte(identifier))] = frost.NewPackage(pkg)
		}

		r2Packages, r2SecretPtr, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Pkgs)
		if err != nil {
			return
		}
		e.artifacts.r2 = &Round2Result{
			pkgs:   r2Packages,
			secret: NewSecret(r2SecretPtr),
		}
	}

	for identifierTo := range e.artifacts.r2.pkgs {
		e.coordinatorContract.SendRound2(
			validatorIdx,
			localIdentifier,
			identifierTo.ToBytes(),
			e.artifacts.r2.pkgs[identifierTo].ToBytes(),
		)
	}
}

func (e *Executor) executeR3(dkg *coordinatorcontract.DKG, validatorIdx uint16, localIdentifier []byte) {
	e.logDKGProcess(dkg, "Start executing R3")
	if dkg.Round3Completed() {
		e.logDKGProcess(dkg, "R3 completed")
		return
	}
	if !dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 not yet completed, waiting for more packages.")
		return
	}

	if e.artifacts.r3 == nil {

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

		keyPackage, publicKeyPackage, err := frost.DkgPart3(e.artifacts.r2.secret.ptr, r1Packages, r2Packages)
		if err != nil {
			return
		}
		e.artifacts.r3 = &Round3Result{
			pkg:              keyPackage,
			publicKeyPackage: publicKeyPackage,
		}
	}

	e.coordinatorContract.SendPubkeyPackage(
		validatorIdx,
		e.artifacts.r3.pkg,
		localIdentifier,
		e.artifacts.r3.publicKeyPackage,
	)
}
