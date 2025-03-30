package dkg

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
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
	pkgs                  map[frost.Identifier]frost.Package
	secret                Secret
	maliciousValidatorIdx []byte
}

type Round3Result struct {
	secretPackage         []byte
	publicKeyPackage      []byte
	publicKey             []byte // 33 bytes with prefix
	maliciousValidatorIdx []byte
}

type ExecutionArtifacts struct {
	r1 *Round1Result
	r2 *Round2Result
	r3 *Round3Result
}

func (a *ExecutionArtifacts) Cleanup() {
	if a.r2 != nil && a.r2.secret.ptr != 0 {
		frost.FreeR2Secret(a.r2.secret.ptr)
	}
	a.r1 = nil
	a.r2 = nil
	a.r3 = nil
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
		e.artifacts.Cleanup()
		e.validator.GetSessionSigner().OnNewDKG(dkg.Until.Unix())
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
	sessionPublicKey := e.validator.GetSessionSigner().PublicKey()

	if e.executeR1(dkg, keyInfo.VsetIdx, sessionPublicKey) {
		if e.executeR2(dkg, keyInfo.VsetIdx) {
			if e.executeClaimR2(dkg, keyInfo.VsetIdx) {
				if e.executeR3(dkg, keyInfo.VsetIdx) {
					if e.executeClaimR3(dkg, keyInfo.VsetIdx) {
						e.logDKGProcess(dkg, "Successfully completed all DKG rounds")
					}
				}
			}
		}
	}
}

func (e *Executor) executeR1(dkg *coordinator.DKG, validatorIdx uint16, sessionPublicKey []byte) bool {
	e.logExecuteR1(dkg)
	if dkg.Round1Completed() {
		e.logDKGProcess(dkg, "R1 completed")
		return true
	}

	packages := dkg.GetR1Packages()
	localIdentifier := helpers.ValidatorIdxToFrost(validatorIdx)
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
		sessionPublicKey,
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
		delete(r1Packages, frost.Identifier(localIdentifier))

		r2Packages, r2SecretPtr, maliciousValidatorIdx, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Packages)
		if err != nil {
			if maliciousValidatorIdx != nil {
				e.logError(dkg, "Part2 failed. Malicious validator found.", err)

				//e.artifacts.r2 = &Round2Result{
				//	pkgs:                  nil,
				//	secret:                NewSecret(0),
				//	maliciousValidatorIdx: maliciousValidatorIdx,
				//}
			} else {
				e.logError(dkg, "Part2 failed", err)
			}
			return false
		}
		e.artifacts.r2 = &Round2Result{
			pkgs:                  r2Packages,
			secret:                NewSecret(r2SecretPtr),
			maliciousValidatorIdx: nil,
		}
	}

	e.logMessage(dkg, "sending r2 packages...")
	withErrors := false
	// Get r2 packages that are already sent to coordinator.
	// But only from this oracle to others
	alreadySentPackages := dkg.GetR2Packages(localIdentifier)
	// Go through all r2 packages generated locally
	for identifierTo, r2pkg := range e.artifacts.r2.pkgs {
		// Check if oracle has already sent this package to coordinator
		_, exists := alreadySentPackages[hex.EncodeToString(identifierTo.ToBytes())]
		if !exists {
			// local r2 package is not sent yet, send it to coordinator
			_, err := e.coordinatorContract.SendRound2(
				validatorIdx,
				localIdentifier,
				identifierTo.ToBytes(),
				r2pkg.ToBytes(),
			)
			if err != nil {
				e.logSendRound2Package(dkg, identifierTo.ToBytes(), err)
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

func (e *Executor) executeClaimR2(dkg *coordinator.DKG, validatorIdx uint16) bool {
	e.logExecuteR2Claim(dkg)
	if dkg.ClaimCompleted(validatorIdx) {
		e.logDKGProcess(dkg, "R2 claim completed")
		return true
	}

	if e.artifacts.r2 == nil {
		e.logError(dkg, "Round2 artifacts not found", nil)
		return false
	}

	maliciousValidatorIdxStr := "NO"
	if e.artifacts.r2.maliciousValidatorIdx != nil {
		maliciousValidatorIdxStr = hex.EncodeToString(e.artifacts.r2.maliciousValidatorIdx)
	}
	e.logMessage(dkg, "sending r2 claim packages. Malicious validator idx: "+maliciousValidatorIdxStr)
	withErrors := false

	// r2 claim package is not sent yet, send it to coordinator
	_, err := e.coordinatorContract.SendClaim(
		validatorIdx,
		e.artifacts.r2.maliciousValidatorIdx,
	)
	if err != nil {
		e.logSendClaimPackage(dkg, e.artifacts.r2.maliciousValidatorIdx, err)
		withErrors = true
	}

	if withErrors {
		e.logError(dkg, "R2 claim packages sent with errors", nil)
	} else {
		e.logDKGProcess(dkg, "R2 claim packages sent")
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
		delete(r1Packages, frost.Identifier(localIdentifier))

		r2Packages := helpers.ConvertMapToFrostPackages(dkg.GetR2Packages(localIdentifier))

		keyPackage, publicKeyPackage, maliciousValidatorIdx, err := frost.DkgPart3(e.artifacts.r2.secret.ptr, r1Packages, r2Packages)
		if err != nil {
			if maliciousValidatorIdx != nil {
				e.logError(dkg, "Part2 failed. Malicious validator found.", err)

				e.artifacts.r3 = &Round3Result{
					secretPackage:         nil,
					publicKeyPackage:      nil,
					publicKey:             nil,
					maliciousValidatorIdx: maliciousValidatorIdx,
				}
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
			secretPackage:         keyPackage,
			publicKeyPackage:      publicKeyPackage,
			publicKey:             publicKey,
			maliciousValidatorIdx: nil,
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

func (e *Executor) executeClaimR3(dkg *coordinator.DKG, validatorIdx uint16) bool {
	e.logExecuteR3Claim(dkg)
	if dkg.ClaimCompleted(validatorIdx) {
		e.logDKGProcess(dkg, "R3 claim completed")
		return true
	}

	if e.artifacts.r3 == nil {
		e.logError(dkg, "Round3 artifacts not found", nil)
		return false
	}

	maliciousValidatorIdxStr := "NO"
	if e.artifacts.r3.maliciousValidatorIdx != nil {
		maliciousValidatorIdxStr = hex.EncodeToString(e.artifacts.r3.maliciousValidatorIdx)
	}
	e.logMessage(dkg, "sending r3 claim packages. Malicious validator idx: "+maliciousValidatorIdxStr)
	withErrors := false

	// r3 claim package is not sent yet, send it to coordinator
	_, err := e.coordinatorContract.SendClaim(
		validatorIdx,
		e.artifacts.r3.maliciousValidatorIdx,
	)
	if err != nil {
		e.logSendClaimPackage(dkg, e.artifacts.r3.maliciousValidatorIdx, err)
		withErrors = true
	}

	if withErrors {
		e.logError(dkg, "R3 claim packages sent with errors", nil)
	} else {
		e.logDKGProcess(dkg, "R3 claim packages sent")
	}
	return false
}
