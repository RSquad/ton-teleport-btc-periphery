package dkg

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
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
	secretPackage    []byte
	publicKeyPackage []byte
	publicKey        []byte // 33 bytes with prefix
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

func (e *Executor) Work(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dkg, ok := <-e.inChan:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			e.Execute(dkg)
		}
	}
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
		e.artifacts = ExecutionArtifacts{}
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
	e.logExecuteR1(dkg)
	if dkg.Round1Completed() {
		e.logDKGProcess(dkg, "R1 completed")
		return true
	}

	packages := dkg.GetR1Packages()
	if packages[hex.EncodeToString(localIdentifier)] != nil {
		e.logDKGProcess(dkg, "R1 package already stored in DKG")
		return false
	}

	if e.artifacts.r1 == nil {
		e.logMessage(dkg, "generating round1 artifacts")
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
		localIdentifier, // sessionPublicKey
	)
	if err != nil {
		e.logSendRound1Package(dkg, err)
	} else {
		e.logDKGProcess(dkg, "R1 package sent")
	}
	return false
}

func (e *Executor) executeR2(dkg *coordinator.DKG, validatorIdx uint16, localIdentifier []byte) bool {
	e.logExecuteR2(dkg)
	if dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 completed")
		return true
	}

	if e.artifacts.r2 == nil {
		if e.artifacts.r1 == nil {
			e.logError(dkg, "Round1 artifacts not found", nil)
			return false
		}

		e.logMessage(dkg, "generating round2 artifacts")
		r1Packages := helpers.ConvertMapToFrostPackages(dkg.GetR1Packages())
		delete(r1Packages, frost.Identifier(localIdentifier))

		r2Packages, r2SecretPtr, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Packages)
		if err != nil {
			e.logError(dkg, "Part2 failed", err)
			return false
		}
		e.artifacts.r2 = &Round2Result{
			pkgs:   r2Packages,
			secret: NewSecret(r2SecretPtr),
		}
	}

	withErrors := false
	r2Packages := dkg.GetR2Packages(localIdentifier)
	for identifierTo := range e.artifacts.r2.pkgs {
		_, exists := r2Packages[hex.EncodeToString(identifierTo.ToBytes())]
		if !exists {
			_, err := e.coordinatorContract.SendRound2(
				validatorIdx,
				localIdentifier,
				identifierTo.ToBytes(),
				e.artifacts.r2.pkgs[identifierTo].ToBytes(),
			)
			if err != nil {
				e.logSendRound2Package(dkg, err)
				withErrors = true
			}
		}
	}
	if withErrors {
		e.logError(dkg, "R2 packages sent with errors", nil)
	} else {
		e.logDKGProcess(dkg, "R2 packages sent")
	}
	return false
}

func (e *Executor) executeR3(dkg *coordinator.DKG, validatorIdx uint16, localIdentifier []byte) bool {
	e.logExecuteR3(dkg)
	if dkg.Round3Completed() {
		e.logDKGProcess(dkg, "R3 completed")
		return true
	}
	if !dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 not yet completed, waiting for more packages.")
		return false
	}

	if e.artifacts.r3 == nil {
		if e.artifacts.r2 == nil {
			e.logError(dkg, "Round2 artifacts not found", nil)
			return false
		}

		r1Packages := helpers.ConvertMapToFrostPackages(dkg.GetR1Packages())
		delete(r1Packages, frost.Identifier(localIdentifier))

		r2Packages := helpers.ConvertMapToFrostPackages(dkg.GetR2Packages(localIdentifier))
		delete(r2Packages, frost.Identifier(localIdentifier))

		keyPackage, publicKeyPackage, err := frost.DkgPart3(e.artifacts.r2.secret.ptr, r1Packages, r2Packages)
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("R3 failed: %v", err))
			return false
		}

		publicKey, err := frost.ExtractPublicKeyFromPackage(publicKeyPackage)
		if err != nil {
			e.logError(dkg, "failed to extract public key from package", err)
			return false
		}
		e.artifacts.r3 = &Round3Result{
			secretPackage:    keyPackage,
			publicKeyPackage: publicKeyPackage,
			publicKey:        publicKey,
		}
		err = e.keystore.StoreSecret(publicKey[1:], keyPackage)
		if err != nil {
			e.logError(dkg, "failed to store secret", err)
			return false
		}
	}

	if _, err := e.coordinatorContract.SendPubkeyPackage(
		validatorIdx,
		e.artifacts.r3.publicKey[1:], // skip prefix byte
		localIdentifier,
		e.artifacts.r3.publicKeyPackage,
	); err != nil {
		e.logSendPubkeyPackageFailed(dkg, err)
	}
	return false
}
