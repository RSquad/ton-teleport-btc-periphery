package dkg

import (
	"context"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
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
	inChan              chan *coordinator.DKG
	until               time.Time
	coordinatorContract *coordinator.CoordinatorContract
	artifacts           ExecutionArtifacts
	keystore            keystore.Keystore
	validator           *validator.Validator
}

func NewExecutor(
	inChan chan *coordinator.DKG,
	coordinatorContract *coordinator.CoordinatorContract,
	keystore keystore.Keystore,
	validator *validator.Validator,
) *Executor {
	return &Executor{
		inChan:              inChan,
		until:               time.Unix(0, 0),
		coordinatorContract: coordinatorContract,
		artifacts:           ExecutionArtifacts{},
		keystore:            keystore,
		validator:           validator,
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

func (e *Executor) startDkgServer(ctx context.Context, endpoint *Endpoint) (err error) {
	logger.DefaultLogStartWork("DKGServer")
	defer logger.DefaultLogFinishWork("DKGServer", err)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case request, ok := <-endpoint.CommitRequestCh:
			if !ok {
				continue
			}
			nonce, commitments, err := e.Commit(request.publicKey)
			if err != nil {
				return err
			}
			endpoint.CommitResultCh <- &CommitResult{
				Nonce:       nonce,
				Commitments: commitments,
			}

		case request, ok := <-endpoint.SignRequestCh:
			if !ok {
				continue
			}
			signingShare, err := e.Sign(request.internalKey, request.tapTweak,
				request.commitments, request.nonceName, request.message)
			if err != nil {
				return err
			}
			endpoint.SignResultCh <- &SignResult{
				signingShare: signingShare,
			}
		}
	}
}

func (e *Executor) Sign(
	publicKey []byte,
	tapTweak []byte,
	commitments map[string][]byte,
	nonceName string,
	message []byte,
) ([]byte, error) {
	secret := e.keystore.LoadSecret(publicKey)
	if secret == nil {
		return nil, fmt.Errorf("failed to load secret by key %x", publicKey)
	}
	nonce := e.keystore.LoadNonce(nonceName)
	if nonce == nil {
		return nil, fmt.Errorf("failed to load nonce by name %s", nonceName)
	}
	frostCommitments := helpers.ConvertMapToFrostPackages(commitments)
	return frost.SignWithTweak(frost.NewPackage(secret), message, frostCommitments, frost.NewPackage(nonce))
}

func (e *Executor) Commit(publicKey []byte) ([]byte, []byte, error) {
	secret := e.keystore.LoadSecret(publicKey)
	if secret == nil {
		return nil, nil, fmt.Errorf("failed to load secret by key %x", publicKey)
	}
	return frost.Commit(frost.NewPackage(secret))
}

func (e *Executor) Execute(dkg *coordinator.DKG) {
	e.logStartExecuting(dkg)
	defer e.logFinishExecuting(dkg)

	if dkg.Status == coordinator.DKGStatusFinished {
		e.logDKGFinished(dkg)
		return
	}

	if dkg.Until.After(e.until) {
		e.until = dkg.Until
		e.logNewDKGStarted(dkg)
	}

	keyInfo, err := e.validator.FindKeyInfo(dkg.VSet)
	if err != nil {
		e.logDKGProcess(dkg, fmt.Sprintf("Error finding key info: %v", err))
		return
	}
	if keyInfo == nil {
		e.logDKGProcess(dkg, "Oracle is not a future validator. Cannot participate in DKG.")
		return
	}
	e.coordinatorContract.ConnectSigner(e.validator.GetSigner(keyInfo.KeyID))

	if e.executeR1(dkg, keyInfo.VsetIdx, keyInfo.PublicKey) {
		if e.executeR2(dkg, keyInfo.VsetIdx, keyInfo.PublicKey) {
			if e.executeR3(dkg, keyInfo.VsetIdx, keyInfo.PublicKey) {
				e.logDKGProcess(dkg, "Successfully completed all DKG rounds")
			}
		}
	}
}

func (e *Executor) executeR1(dkg *coordinator.DKG, validatorIdx uint16, localIdentifier []byte) bool {
	e.logDKGProcess(dkg, "Start executing R1")
	if dkg.Round1Completed() {
		e.logDKGProcess(dkg, "R1 completed")
		return true
	}

	packages := dkg.GetR1Packages()
	if packages[string(localIdentifier)] != nil {
		e.logDKGProcess(dkg, "R1 package already sent")
		return false
	}

	if e.artifacts.r1 == nil {
		minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("Failed to calculate min signers: %v", err))
			return false
		}

		r1Package, r1SecretPtr, err := frost.DkgPart1(
			localIdentifier,
			minSigners,
			dkg.MaxSigners,
		)
		if err != nil {
			e.logDKGPart1Failed(dkg, err)
			return false
		}
		e.artifacts.r1 = &Round1Result{
			pkg:    r1Package,
			secret: NewSecret(r1SecretPtr),
		}
	}

	_, err := e.coordinatorContract.SendRound1(
		validatorIdx,
		localIdentifier,
		e.artifacts.r1.pkg,
	)
	if err != nil {
		msg := helpers.HandleTvmError(err)
		e.logDKGProcess(dkg, "Failed to send r1 package: "+msg)
		return false
	} else {
		e.logDKGProcess(dkg, "R1 package sent")
	}
	return false
}

func (e *Executor) executeR2(dkg *coordinator.DKG, validatorIdx uint16, localIdentifier []byte) bool {
	e.logDKGProcess(dkg, "Start executing R2")
	if dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 completed")
		return true
	}

	if e.artifacts.r2 == nil {
		r1Pkgs := make(map[frost.Identifier]frost.Package)

		for identifier, pkg := range dkg.GetR1Packages() {
			if identifier == string(localIdentifier) {
				continue
			}

			r1Pkgs[frost.Identifier([]byte(identifier))] = frost.NewPackage(pkg)
		}

		r2Packages, r2SecretPtr, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Pkgs)
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("Part2 failed: %v", err))
			return false
		}
		e.artifacts.r2 = &Round2Result{
			pkgs:   r2Packages,
			secret: NewSecret(r2SecretPtr),
		}
	}

	for identifierTo := range e.artifacts.r2.pkgs {
		_, err := e.coordinatorContract.SendRound2(
			validatorIdx,
			localIdentifier,
			identifierTo.ToBytes(),
			e.artifacts.r2.pkgs[identifierTo].ToBytes(),
		)
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("Failed to send R2 package: %v", err))
			break
		}
	}
	return false
}

func (e *Executor) executeR3(dkg *coordinator.DKG, validatorIdx uint16, localIdentifier []byte) bool {
	e.logDKGProcess(dkg, "Start executing R3")
	if dkg.Round3Completed() {
		e.logDKGProcess(dkg, "R3 completed")
		return true
	}
	if !dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 not yet completed, waiting for more packages.")
		return false
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
			e.logDKGProcess(dkg, fmt.Sprintf("R3 failed: %v", err))
			return false
		}
		e.artifacts.r3 = &Round3Result{
			pkg:              keyPackage,
			publicKeyPackage: publicKeyPackage,
		}
	}

	_, err := e.coordinatorContract.SendPubkeyPackage(
		validatorIdx,
		e.artifacts.r3.pkg,
		localIdentifier,
		e.artifacts.r3.publicKeyPackage,
	)
	if err != nil {
		e.logDKGProcess(dkg, fmt.Sprintf("Failed to send pubkey package: %v", err))
	}
	return false
}
