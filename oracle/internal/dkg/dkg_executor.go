package dkg

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
	"golang.org/x/crypto/nacl/box"
)

type Secret struct {
	ptr uintptr
}

func NewSecret(ptr uintptr) Secret {
	return Secret{ptr: ptr}
}

type Round1Result struct {
	pkg             []byte
	secret          Secret
	r2PublicX25519  *[32]byte
	r2PrivateX25519 *[32]byte
}

type Round2Result struct {
	packages []byte // Encrypted and serialized packages
	secret   Secret
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
	a.SafeCleanPrivateX25519()
	if a.r2 != nil && a.r2.secret.ptr != 0 {
		frost.FreeR2Secret(a.r2.secret.ptr)
	}
	a.r1 = nil
	a.r2 = nil
	a.r3 = nil
}

func (a *ExecutionArtifacts) SafeCleanPrivateX25519() {
	if a.r1 == nil {
		return
	}

	*a.r1.r2PrivateX25519 = [32]byte{}
}

func (a *ExecutionArtifacts) IsEmpty() bool {
	return a.r1 == nil && a.r2 == nil && a.r3 == nil
}

type Executor struct {
	inChan              chan *coordinator.DKG
	until               time.Time
	coordinatorContract coordinator.Coordinator
	artifacts           ExecutionArtifacts
	keystore            keystore.Keystore
	validator           *validator.Validator
	sessionSigner       signer.Signer
	validatorIdx        uint16
	cfg                 *cfg.Cfg
}

func NewExecutor(
	inChan chan *coordinator.DKG,
	coordinatorContract coordinator.Coordinator,
	keystore keystore.Keystore,
	validator *validator.Validator,
	cfg *cfg.Cfg,
) *Executor {
	return &Executor{
		inChan:              inChan,
		until:               time.Unix(0, 0),
		coordinatorContract: coordinatorContract,
		artifacts:           ExecutionArtifacts{},
		keystore:            keystore,
		validator:           validator,
		sessionSigner:       nil,
		validatorIdx:        255,
		cfg:                 cfg,
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
	e.sessionSigner = nil
}

func (e *Executor) OnStartNewDKG(dkg *coordinator.DKG) bool {
	e.Cleanup()

	// Get session public key

	sessionSigner, err := validator.NewSessionSigner(e.keystore, dkg.Until.Unix())
	if err != nil {
		e.logDKGProcess(dkg, fmt.Sprintf("Failed to create SessionSigner: %v", err))
		return false
	}
	e.sessionSigner = sessionSigner

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

	e.validatorIdx = keyInfo.VsetIdx

	if dkg.Claims.Count > 0 {
		e.logDKGClaims(dkg)
	}

	if !dkg.CheckVSetMask(e.validatorIdx) {
		e.logDKGProcess(dkg, "The Oracle has been EVICTED from DKG")
		return
	}

	e.coordinatorContract.ConnectSigner(e.validator.GetSigner(keyInfo.KeyID))

	if e.executeR1(dkg) {
		if e.executeR2(dkg) {
			if e.executeR3(dkg) {
				e.logDKGProcess(dkg, "Successfully completed all DKG rounds!")
			}
		}
	}
}

func (e *Executor) executeR1(dkg *coordinator.DKG) bool {
	e.logExecuteR1(dkg)
	if dkg.Round1Completed() {
		e.logDKGR1Completed(dkg)
		return true
	}

	// Check R1 mask
	if r, cnt := dkg.CheckR1Mask(e.validatorIdx); r {
		e.logDKGProcess(dkg, fmt.Sprintf("R1 package already stored in DKG. Waiting for other Oracles (ready %d of %d)", cnt, dkg.MaxSigners))
		return false
	}

	if e.cfg.TestSkipR1 {
		e.logDebug("Skip R1 stage")
		return false
	}

	if e.artifacts.r1 == nil {
		e.logMessage(dkg, "Generating R1 artifacts")
		minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
		if err != nil {
			e.logDKGProcess(dkg, fmt.Sprintf("Failed to calculate min signers: %v", err))
			return false
		}

		if e.cfg.TestInvalidSigners {
			e.logDebug("Set minSigners = maxSigners")
			minSigners = dkg.MaxSigners
		}

		r1Package, r1SecretPtr, err := frost.DkgPart1(
			helpers.ValidatorIdxToFrost(e.validatorIdx),
			minSigners,
			dkg.MaxSigners,
		)
		if err != nil {
			e.logDKGPart1Failed(dkg, err)
			return false
		}

		if e.cfg.TestBadR1Pkg {
			e.logDebug("Generate random R1 package")
			randomBytes := make([]byte, len(r1Package))
			_, err := rand.Read(randomBytes)
			if err != nil {
				return false
			}
			r1Package = randomBytes
			if !e.cfg.TestBadR1PkgRandomVersion {
				e.logDebug("R1 package version is 0")
				r1Package[0] = 0
			}
		}

		// Generate key pair for Round2 encryption
		r2PublicX25519, r2PrivateX25519, err := box.GenerateKey(rand.Reader)
		if err != nil {
			e.logDKGPart1Failed(dkg, err)
			return false
		}

		e.artifacts.r1 = &Round1Result{
			pkg:             r1Package,
			secret:          NewSecret(r1SecretPtr),
			r2PublicX25519:  r2PublicX25519,
			r2PrivateX25519: r2PrivateX25519,
		}
	}

	_, err := e.coordinatorContract.SendRound1(
		e.validatorIdx,
		dkg.Until.Unix(),
		e.artifacts.r1.pkg,
		e.artifacts.r1.r2PublicX25519,
	)
	if err != nil {
		e.logSendRound1Package(dkg, err)
	} else {
		e.logDKGProcess(dkg, "R1 package sent")
	}
	return false
}

func (e *Executor) executeR2(dkg *coordinator.DKG) bool {
	e.logExecuteR2(dkg)
	if dkg.Round2Completed() {
		e.logDKGR2Completed(dkg)
		return true
	}

	// Check R2 mask
	if r, cnt := dkg.CheckR2Mask(e.validatorIdx); r {
		e.logDKGProcess(dkg, fmt.Sprintf("R2 packages already stored in DKG. Waiting for other Oracles (ready %d of %d)", cnt, dkg.MaxSigners))
		return false
	}

	if e.cfg.TestSkipR2 {
		e.logDebug("Skip R2 stage")
		return true
	}

	localIdentifier := helpers.ValidatorIdxToFrost(e.validatorIdx)

	if e.artifacts.r2 == nil {
		if e.artifacts.r1 == nil {
			e.logError(dkg, "R1 artifacts not found. Waiting for DKG to restart.", nil)
			return false
		}

		e.logMessage(dkg, "Generating R2 artifacts")
		r1Packages, r2PublicKeysX25519, culpritIdx, err := helpers.DeserializeDkgR1(dkg.GetR1Packages())
		if err != nil {
			e.logError(dkg, "Failed to parse R1 packages. Culprit validator found.", err)
			e.executeClaim(dkg, culpritIdx)
			return false
		}
		delete(r1Packages, localIdentifier)

		r2Packages, r2SecretPtr, culpritFrostIdx, err := frost.DkgPart2(e.artifacts.r1.secret.ptr, r1Packages)
		if err != nil {
			if culpritFrostIdx != nil {
				culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
				e.logError(dkg, fmt.Sprintf("R2 failed. Culprit validator found: %d", culpritIdx), err)
				e.executeClaim(dkg, culpritIdx)
			} else {
				e.logError(dkg, "R2 failed", err)
			}
			return false
		}

		if e.cfg.TestBadR2Pkg {
			e.logDebug("Generate random R2 package")
			newR2Packages := make(map[frost.Identifier]frost.Package)
			for id, pkg := range r2Packages {
				randomBytes := make([]byte, len(pkg.ToBytes()))
				_, err := rand.Read(randomBytes)
				if err != nil {
					return false
				}
				newR2Packages[id] = frost.NewPackage(randomBytes)
			}
			r2Packages = newR2Packages
		}

		// Encrypt R2 packages
		r2EncryptedPackages, err := EncryptR2Packages(
			r2Packages,
			r2PublicKeysX25519,
			e.artifacts.r1.r2PrivateX25519,
			dkg.Until,
			e.validatorIdx,
		)
		if err != nil {
			e.logError(dkg, "Failed to encrypt R2 packages", err)
			return false
		}

		// Convert R2 packages to bytes
		r2packagesSerialized := helpers.SerializeR2Packages(r2EncryptedPackages)

		if e.cfg.TestBadR2Serialized {
			e.logDebug("Damage serialized r2 packages")
			r2packagesSerialized = make([]byte, len(r2packagesSerialized))
			_, err := rand.Read(r2packagesSerialized)
			if err != nil {
				return false
			}
		}

		// Save the result into the artifacts
		e.artifacts.r2 = &Round2Result{
			packages: r2packagesSerialized,
			secret:   NewSecret(r2SecretPtr),
		}
	}

	// R2 package is not sent yet, send it to coordinator
	e.logMessage(dkg, "Sending R2 package...")

	_, err := e.coordinatorContract.SendRound2(
		e.validatorIdx,
		dkg.Until.Unix(),
		e.artifacts.r2.packages,
	)

	if err != nil {
		e.logSendRound2Package(dkg, err)
	} else {
		e.logDKGProcess(dkg, "R2 packages sent")
	}
	return false
}

func (e *Executor) executeR3(dkg *coordinator.DKG) bool {
	e.logExecuteR3(dkg)

	if dkg.Round3Completed() {
		e.artifacts.SafeCleanPrivateX25519()
		e.logDKGR3Completed(dkg)
		return true
	}

	// Check R3 mask
	if r, cnt := dkg.CheckR3Mask(e.validatorIdx); r {
		e.logDKGProcess(dkg, fmt.Sprintf("R3 packages already stored in DKG. Waiting for other Oracles (ready %d of %d)", cnt, dkg.MaxSigners))
		return false
	}

	if e.cfg.TestSkipR3 {
		e.logDebug("Skip R3 stage")
		return true
	}

	localIdentifier := helpers.ValidatorIdxToFrost(e.validatorIdx)

	if e.artifacts.r3 == nil {
		if e.artifacts.r2 == nil {
			e.logError(dkg, "R2 artifacts not found. Waiting for DKG to restart", nil)
			return false
		}

		// Get R1 packages
		r1Packages, r2PublicKeysX25519, culpritIdx, err := helpers.DeserializeDkgR1(dkg.GetR1Packages())
		if err != nil {
			e.logError(dkg, "Failed to parse R1 packages. Culprit validator found.", err)
			e.executeClaim(dkg, culpritIdx)
			return false
		}
		delete(r1Packages, localIdentifier)

		// Get R2 packages
		r2Packages, isCulpritFound, culpritIdx, err := helpers.DeserializeDkgR2(
			dkg.GetR2Packages(),
			dkg.VSetMask,
		)

		if isCulpritFound {
			e.logError(dkg, "R3 failed. Failed to parse R2 packages. Culprit validator found.", err)
			e.executeClaim(dkg, culpritIdx)
			return false
		}

		if err != nil {
			e.logError(dkg, "R3 failed. An error occurred while trying to get R2 packages", err)
			return false
		}

		// Decrypt R2 packages
		r2PackagesDecrypted, isCulpritFound, culpritIdx, err := DecryptR2Packages(
			r2Packages,
			e.validatorIdx,
			r2PublicKeysX25519,
			e.artifacts.r1.r2PrivateX25519,
			dkg.Until,
		)

		if isCulpritFound {
			e.logError(dkg, "R3 failed. Failed to decrypt R2 packages. Culprit validator found.", err)
			e.executeClaim(dkg, culpritIdx)
			return false
		}

		if err != nil {
			e.logError(dkg, "R3 failed. An error occurred while trying to decrypt R2 packages", err)
			return false
		}

		keyPackage, publicKeyPackage, culpritFrostIdx, err := frost.DkgPart3(e.artifacts.r2.secret.ptr, r1Packages, r2PackagesDecrypted)
		if err != nil {
			if culpritFrostIdx != nil {
				e.logError(dkg, "Part3 failed. Culprit validator found.", err)
				culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
				e.executeClaim(dkg, culpritIdx)
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

		if e.cfg.TestBadR3Pkg {
			e.logDebug("Generate random R3 package")
			randomBytes := make([]byte, len(publicKeyPackage))
			_, err := rand.Read(randomBytes)
			if err != nil {
				return false
			}
			publicKeyPackage = randomBytes
		}

		e.artifacts.r3 = &Round3Result{
			secretPackage:    keyPackage,
			publicKeyPackage: publicKeyPackage,
			publicKey:        publicKey,
		}
		e.artifacts.SafeCleanPrivateX25519()
		err = e.keystore.StoreSecret(publicKey[1:], keyPackage)
		if err != nil {
			e.logError(dkg, "failed to store secret", err)
			return false
		}
	}

	if _, err := e.coordinatorContract.SendPubkeyPackage(
		e.validatorIdx,
		dkg.Until.Unix(),
		e.sessionSigner,
		e.artifacts.r3.publicKeyPackage,
	); err != nil {
		exitCode, _ := helpers.ExtractExitCode(err.Error())
		if exitCode == helpers.TvmExitCodeDifferentPubkeyPackages {
			e.claimCulpritByR3Mask(dkg)
		}
		e.logSendPubkeyPackageFailed(dkg, err)
	} else {
		e.logDKGProcess(dkg, "R3 packages sent")
	}

	return false
}

func (e *Executor) executeClaim(dkg *coordinator.DKG, culpritIdx uint16) {
	e.logExecuteClaim(dkg)

	if dkg.ClaimCompleted(e.validatorIdx) {
		e.logDKGProcess(dkg, "claim completed")
		return
	}

	e.logMessage(dkg, fmt.Sprintf("sending claim packages. Culprit validator idx: %d", culpritIdx))
	_, err := e.coordinatorContract.SendDKGClaim(
		e.validatorIdx,
		dkg.Until.Unix(),
		culpritIdx,
	)
	if err != nil {
		e.logSendClaimFailed(dkg, culpritIdx, err)
	} else {
		e.logDKGProcess(dkg, "claim packages sent")
	}
}

func (e *Executor) claimCulpritByR3Mask(dkg *coordinator.DKG) {
	for i := uint16(0); i < dkg.MaxSigners; i++ {
		if dkg.R3.Mask.Bit(int(i)) > 0 {
			e.executeClaim(dkg, i)
			return
		}
	}
}
