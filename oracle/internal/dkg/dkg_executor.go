package dkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
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

func (a *ExecutionArtifacts) Cleanup() {
	if a.r1 != nil && a.r1.secret.ptr != 0 {
		frost.FreeR1Secret(a.r1.secret.ptr)
	}
	if a.r2 != nil && a.r2.secret.ptr != 0 {
		frost.FreeR2Secret(a.r2.secret.ptr)
	}
	a.r1 = nil
	a.r2 = nil
	a.r3 = nil
}

func (a *ExecutionArtifacts) IsEmpty() bool {
	return a.r1 == nil && a.r2 == nil && a.r3 == nil
}

type Executor struct {
	inChan              chan *coordinator.DKG
	until               time.Time
	coordinatorContract *coordinator.CoordinatorContract
	artifacts           ExecutionArtifacts
	keystore            keystore.Keystore
	validator           *validator.Validator
	sessionPublicKey    []byte
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

func (e *Executor) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.DefaultLogFinishWork("DKG Executor")
	logger.DefaultLogStartWork("DKG Executor")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Executor received shutdown signal...")
			return
		case dkg, ok := <-e.inChan:
			if !ok {
				logger.Log.Warn().Msg("DKG Executor channel closed")
				return
			}
			e.Execute(dkg)
		}
	}
}

func (e *Executor) Cleanup() {
	e.artifacts.Cleanup()
	e.sessionPublicKey = nil
}

func (e *Executor) OnStartNewDKG(dkg *coordinator.DKG) bool {
	e.Cleanup()

	// Get session public key
	{
		sessionSigner, err := validator.NewSessionSigner(e.keystore, dkg.Until.Unix())
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("Failed to create SessionSigner: %v", err))
			return false
		}
		e.sessionPublicKey = sessionSigner.PublicKey()
	}

	e.until = dkg.Until
	e.logNewDKGStarted(dkg)
	return true
}

func (e *Executor) Execute(dkg *coordinator.DKG) {
	e.logStartExecuting(dkg)
	defer e.logFinishExecuting(dkg)

	// Verify if DKG is finished
	if dkg.State == coordinator.DKGStateFinished {
		e.Cleanup()
		e.logDKGFinished(dkg)
		return
	}

	// Verify if it needs to start a new DKG
	if dkg.Until.After(e.until) {
		if !e.OnStartNewDKG(dkg) {
			return
		}
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

	if !dkg.CheckMask(keyInfo.VsetIdx) {
		e.logDKGProcess(dkg, "The Oracle has been evicted from DKG")
		return
	}

	e.coordinatorContract.ConnectSigner(e.validator.GetSigner(keyInfo.KeyID))

	if e.executeR1(dkg, keyInfo.VsetIdx) {
		if e.executeR2(dkg, keyInfo.VsetIdx) {
			if e.executeR3(dkg, keyInfo.VsetIdx) {
				e.logDKGProcess(dkg, "Successfully completed all DKG rounds")
			}
		}
	}
}

func (e *Executor) executeR1(dkg *coordinator.DKG, validatorIdx uint16) bool {
	e.logExecuteR1(dkg)
	if dkg.Round1Completed() {
		e.logDKGProcess(dkg, "R1 completed")
		return true
	}

	packages := dkg.GetR1Packages()
	if packages[validatorIdx] != nil {
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
			helpers.ValidatorIdxToFrost(validatorIdx),
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
		dkg.Until.Unix(),
		e.artifacts.r1.pkg,
	)
	if err != nil {
		e.logSendRound1Package(dkg, err)
	} else {
		e.logDKGProcess(dkg, "R1 package sent")
	}
	return false
}

func (e *Executor) executeR2(dkg *coordinator.DKG, validatorIdx uint16) bool {
	e.logExecuteR2(dkg)
	if dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 completed")
		return true
	}

	localIdentifier := helpers.ValidatorIdxToFrost(validatorIdx)

	if e.artifacts.r2 == nil {
		if e.artifacts.r1 == nil {
			e.logError(dkg, "Round1 artifacts not found", nil)
			return false
		}

		e.logMessage(dkg, "generating round2 artifacts")
		r1Packages := helpers.ConvertMapToFrostPackages(dkg.GetR1Packages())
		delete(r1Packages, localIdentifier)

		r2Packages, r2SecretPtr, culpritFrostIdx, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Packages)
		if err != nil {
			if culpritFrostIdx != nil {
				culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
				e.logError(dkg, fmt.Sprintf("Part2 failed. Culprit validator found: %d", culpritIdx), err)
				e.executeClaim(dkg, validatorIdx, culpritIdx)
			} else {
				e.logError(dkg, "Part2 failed", err)
			}
			return false
		}

		e.artifacts.r2 = &Round2Result{
			pkgs:   r2Packages,
			secret: NewSecret(r2SecretPtr),
		}
	}

	e.logMessage(dkg, "sending r2 packages...")
	withErrors := false
	// Get r2 packages that are already sent to coordinator.
	// But only from this oracle to others
	// Go through all r2 packages generated locally
	for toIdentificator, r2pkg := range e.artifacts.r2.pkgs {
		// Check if oracle has already sent this package to coordinator

		toIdx := helpers.FrostToValidatorIdx(toIdentificator)
		sentPackages, foundToPackages := dkg.GetR2PackagesTo(toIdx)

		if foundToPackages {
			_, foundFromMePackage := sentPackages[validatorIdx]
			if foundFromMePackage {
				continue
			}
		}

		// local r2 package is not sent yet, send it to coordinator
		_, err := e.coordinatorContract.SendRound2(
			validatorIdx,
			dkg.Until.Unix(),
			toIdx,
			r2pkg.ToBytes(),
		)
		if err != nil {
			e.logSendRound2Package(dkg, toIdx, err)
			withErrors = true
		}
	}

	if withErrors {
		e.logError(dkg, "R2 packages sent with errors", nil)
	} else {
		e.logDKGProcess(dkg, "R2 packages sent")
	}
	return false
}

func (e *Executor) executeR3(dkg *coordinator.DKG, validatorIdx uint16) bool {
	e.logExecuteR3(dkg)

	if dkg.Round3Completed() {
		e.logDKGProcess(dkg, "R3 completed")
		return true
	}

	if !dkg.Round2Completed() {
		e.logDKGProcess(dkg, "R2 not yet completed, waiting for more packages.")
		return false
	}

	localIdentifier := helpers.ValidatorIdxToFrost(validatorIdx)

	if e.artifacts.r3 == nil {
		if e.artifacts.r2 == nil {
			e.logError(dkg, "Round2 artifacts not found", nil)
			return false
		}

		r1Packages := helpers.ConvertMapToFrostPackages(dkg.GetR1Packages())
		delete(r1Packages, localIdentifier)

		sentPackages, foundToPackages := dkg.GetR2PackagesTo(validatorIdx)
		if !foundToPackages {
			e.logError(dkg, "Part3 failed. R2 packages were not found", nil)
			return false
		}

		r2Packages := helpers.ConvertMapToFrostPackages(sentPackages)

		keyPackage, publicKeyPackage, culpritFrostIdx, err := frost.DkgPart3(e.artifacts.r2.secret.ptr, r1Packages, r2Packages)
		if err != nil {
			if culpritFrostIdx != nil {
				e.logError(dkg, "Part3 failed. Culprit validator found.", err)
				culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
				e.executeClaim(dkg, validatorIdx, culpritIdx)
			} else {
				e.logDKGProcess(dkg, fmt.Sprintf("R3 failed: %v", err))
			}
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
		dkg.Until.Unix(),
		e.sessionPublicKey,
		e.artifacts.r3.publicKeyPackage,
	); err != nil {
		e.logSendPubkeyPackageFailed(dkg, err)
	}
	return false
}

func (e *Executor) executeClaim(dkg *coordinator.DKG, validatorIdx uint16, culpritIdx uint16) {
	e.logExecuteClaim(dkg)

	if dkg.ClaimCompleted(validatorIdx) {
		e.logDKGProcess(dkg, "claim completed")
		return
	}

	e.logMessage(dkg, fmt.Sprintf("sending claim packages. Culprit validator idx: %d", culpritIdx))
	withErrors := false

	// claim package is not sent yet, send it to coordinator
	_, err := e.coordinatorContract.SendDKGClaim(
		validatorIdx,
		dkg.Until.Unix(),
		culpritIdx,
	)
	if err != nil {
		e.logSendClaimPackage(dkg, culpritIdx, err)
		withErrors = true
	}

	if withErrors {
		e.logError(dkg, "claim packages sent with errors", nil)
	} else {
		e.logDKGProcess(dkg, "claim packages sent")
	}
}
