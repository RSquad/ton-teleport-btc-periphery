package signing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
	"github.com/xssnick/tonutils-go/ton"
)

type CachedPegout struct {
	ID            uint64
	addrStr       string
	inputs        []pegoutcontract.TxInput
	tx            *pegoutcontract.TxParts
	signingHashes [][]byte
	artifacts     *coordinator.PegoutRecord
}

type SignService struct {
	keyStore          keystore.Keystore
	coordinator       *coordinator.CoordinatorContract
	ton               *tonclient.TonClient
	cachedPegout      *CachedPegout
	executeSignPeriod int64 // `period` in seconds to call the ExecuteSign() function
	dkgUntil          time.Time
	sessionSigner     *validator.SessionSigner
}

func NewService(
	keyStore keystore.Keystore,
	coordinator *coordinator.CoordinatorContract,
	tonclient *tonclient.TonClient,
	executeSignPeriod int64,
) *SignService {
	return &SignService{
		keyStore:          keyStore,
		coordinator:       coordinator,
		ton:               tonclient,
		executeSignPeriod: executeSignPeriod,
		dkgUntil:          time.Unix(0, 0),
		sessionSigner:     nil,
	}
}

func (s *SignService) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.DefaultLogFinishWork("SignService")
	logger.DefaultLogStartWork("SignService")

	// A periodic event that triggers every `period` seconds to call the ExecuteSign() function
	ticker := time.NewTicker(time.Duration(s.executeSignPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Sign service received shutdown signal...")
			return
		case _, ok := <-ticker.C:
			if !ok {
				logger.Log.Warn().Msg("Sign service ticker closed")
				return
			}
			s.ExecuteSign(ctx)
		}
	}
}

func (s *SignService) cachePegoutClear() {
	s.cachedPegout = nil
}

func (s *SignService) cachePegout(
	ctx context.Context,
	unsignedPegout *coordinator.PegoutRecord,
) (*CachedPegout, error) {
	if s.cachedPegout != nil && s.cachedPegout.ID == unsignedPegout.ID {
		s.cachedPegout.artifacts = unsignedPegout
		return s.cachedPegout, nil
	}

	pegoutContract := pegoutcontract.New(unsignedPegout.PegoutAddress, s.ton, ctx)
	block := s.getLatestBlock(ctx)
	txParts, err := pegoutContract.GetTxParts(block)
	if err != nil {
		s.logError("failed to cache pegout", err)
		return nil, err
	}
	txInputs, err := txParts.Inputs.ToSortedSlice()
	if err != nil {
		s.logError("failed to sort inputs", err)
		return nil, err
	}

	signingHashes, err := bitcoin.BuildTaprootSigningHashes(*txParts.Inputs, txParts.Outputs)
	if err != nil {
		s.logError("failed to get signing hashes", err)
		return nil, err
	}

	addrStr := unsignedPegout.PegoutAddress.String()
	s.keyStore.Cleanup()

	s.cachedPegout = &CachedPegout{
		ID:            unsignedPegout.ID,
		addrStr:       addrStr,
		tx:            txParts,
		inputs:        txInputs,
		signingHashes: signingHashes,
		artifacts:     unsignedPegout,
	}
	return s.cachedPegout, nil
}

func (s *SignService) FindValidatorIdx(dkg *coordinator.DKG, sessionPubkey []byte) (uint16, error) {
	for idx, pubkey := range dkg.SessionKeys.PubKeys {
		if bytes.Equal(pubkey, sessionPubkey) {
			return idx, nil
		}
	}

	return 0, errors.New("failed to find validator idx")
}

func (s *SignService) ExecuteSign(ctx context.Context) {
	defer s.logMessage("stop")
	s.logMessage("start")

	dkg, err := s.coordinator.GetPrevDKG()
	if err != nil {
		s.logGetPrevDKGError(err)
		return
	}

	if dkg == nil {
		s.logMessage("previous DKG is not yet initialized")
		return
	}

	s.execute(ctx, dkg)
}

func (s *SignService) execute(ctx context.Context, dkg *coordinator.DKG) {
	pegoutRecords, err := s.coordinator.GetUnsignedPegouts()
	if err != nil {
		s.logUnsignedPegoutsError(err)
		return
	}

	s.logSigningRequestsCount(len(pegoutRecords))

	if len(pegoutRecords) == 0 {
		s.logMessage("No sign requests")
		return
	}

	if (s.dkgUntil != dkg.Until) || (s.sessionSigner == nil) {
		sessionSigner, err := validator.LoadSessionSigner(s.keyStore, dkg.Until.Unix())
		if err != nil {
			s.logError("Failed to create SessionSigner", err)
			return
		}

		s.sessionSigner = sessionSigner
		s.dkgUntil = dkg.Until
	}

	// Get validator idx
	validatorIdx, err := s.FindValidatorIdx(dkg, s.sessionSigner.PublicKey())
	if err != nil {
		s.logError("failed to get validator idx from session key and VSet", err)
		return
	}

	// Get oldest unsigned pegout record
	unsignedPegout := pegoutRecords[0]
	s.logProcessingPegout(&unsignedPegout)

	// Check mask
	if !unsignedPegout.CheckSigningMask(validatorIdx) {
		s.logOracleEvictedFromSigning(unsignedPegout.ID)
		return
	}

	pubkeyPackage := dkg.R3.Data.PubkeyPackage

	// Check for signing restart
	if (unsignedPegout.ExpiredAt != time.Unix(0, 0)) && (unsignedPegout.ExpiredAt.Before(time.Now())) {
		s.executeResetPegoutSigning(unsignedPegout.ID, validatorIdx)
		s.cachePegoutClear()
		return
	}

	// Try caching the pegout
	cachedPegout, err := s.cachePegout(ctx, &unsignedPegout)
	if err != nil {
		s.logError("failed to cache pegout", err)
		return
	}
	if cachedPegout == nil {
		panic("cached pegout is nil")
	}

	s.coordinator.ConnectSigner(s.sessionSigner)

	minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
	if err != nil {
		s.logError("failed to calculate min signers", err)
		return
	}

	// Execute signing steps
	if s.doCommit(validatorIdx, cachedPegout, minSigners) {
		if s.doSign(validatorIdx, cachedPegout, minSigners) {
			if s.doAggregate(validatorIdx, cachedPegout, pubkeyPackage) {
				s.logPegoutSigned(cachedPegout.ID)
			}
		}
	}
}

func (s *SignService) doCommit(
	validatorIdx uint16,
	pegout *CachedPegout,
	minSigners uint16,
) bool {
	s.logCommitPegout(pegout.ID)

	if pegout.artifacts.HasCommitment(validatorIdx) {
		s.logMessage("Commitment already exists")
		if pegout.artifacts.CommitmentsCount() >= minSigners {
			s.logMessage("Commitment round completed")
			return true
		}
		s.logMessage("Waiting for other oracles to commit")
		return false
	}

	nonces := s.keyStore.LoadNonce(pegout.addrStr)
	commitments := s.keyStore.LoadCommitments(pegout.addrStr)

	if nonces == nil || commitments == nil {
		// If both nonce & commitments are not found in keystore
		if nonces == nil && commitments == nil {
			s.logMessage("generate commitments")
			var err error
			nonces, commitments, err = s.Commit(pegout.tx.InternalKey)
			if err != nil {
				s.logError("commit error", err)
				return false
			}
			s.keyStore.StoreNonce(pegout.addrStr, nonces)
			s.keyStore.StoreCommitments(pegout.addrStr, commitments)
		} else {
			s.logErrNullNonceOrCommitments(nonces, commitments, pegout.addrStr)
			return false
		}
	}

	s.logSendCommitments(pegout.ID, commitments)
	if _, err := s.coordinator.SendCommitments(
		pegout.ID,
		validatorIdx,
		commitments,
	); err != nil {
		s.logError("failed to send commitments", err)
	} else {
		s.logCommitSent(pegout.ID)
	}

	return false
}

func (s *SignService) doSign(
	validatorIdx uint16,
	pegout *CachedPegout,
	minSigners uint16,
) bool {
	s.logSignPegout(pegout.ID)

	if !pegout.artifacts.HasCommitment(validatorIdx) {
		s.logErrNoOracleCommitments(pegout.ID)
		return false
	}

	if pegout.artifacts.SigningSharesCount() >= int(minSigners) {
		s.logMinimalSharesReached(pegout.ID)
		return true
	}

	if pegout.artifacts.HasSigningShare(validatorIdx) {
		s.logSigningShareAlreadyExists(pegout.ID)
		return false
	}

	if len(pegout.signingHashes) == 0 {
		s.logErrNothingToSign(pegout.ID)
		return false
	}

	// Check if oracle already generated signing shares
	signShares := s.keyStore.LoadSigningShares(pegout.addrStr)
	if signShares == nil {
		// Share for each signing hash is not generated yet.
		// Call frost.Sign for each signing hash
		s.logMessage("generate signing share")

		// Public key (X coord) used to get secret package from keystore
		publicKey := pegout.tx.InternalKey
		// array with sign shares for each signing hash
		signShares = make([][]byte, 0, len(pegout.signingHashes))
		// generate signing share for each input
		for i, input := range pegout.inputs {
			signShare, culpritFrostIdx, err := s.Sign(
				publicKey,                    // public key
				input.Data.BitcoinMerkleRoot, // tap merkle root used as tweak for tweaking signing share
				pegout.signingHashes[i],      // signing hash for current input
				pegout.artifacts.Commitments, // oracle's commitments to sign pegout hashes
				pegout.addrStr,               // pegout address used as key to load nonce from keystore
			)
			if err != nil {
				if culpritFrostIdx != nil {
					culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
					s.logError(fmt.Sprintf("Sign failed. Culprit validator found: %d", culpritIdx), err)
					s.executeClaim(pegout, validatorIdx, culpritIdx)
				} else {
					s.logError(fmt.Sprintf("failed to sign hash %d", i), err)
				}
				return false
			}

			signShares = append(signShares, signShare)
		}
		s.keyStore.StoreSigningShares(pegout.addrStr, signShares)
	}

	s.logSendSigningShare(pegout.ID, signShares)
	if _, err := s.coordinator.SendSigningShare(
		pegout.ID,
		validatorIdx,
		signShares,
	); err != nil {
		s.logSendSigningShareError(pegout.ID, err)
	} else {
		s.logSigningShareSent(pegout.ID)
	}

	return false
}

func (s *SignService) doAggregate(
	validatorIdx uint16,
	pegout *CachedPegout,
	pubkeyPackage []byte,
) bool {
	s.logAggregateSignShares(pegout.ID)

	commitmentsPackages := helpers.ConvertMapToFrostPackages(pegout.artifacts.Commitments)
	signatures := make([][]byte, 0, len(pegout.signingHashes))

	for i, input := range pegout.inputs {
		hashOnlyShares := filterSharesByHashIndex(pegout.artifacts.SigningShares, uint16(i))
		tapTweak := input.Data.BitcoinMerkleRoot
		signature, culpritFrostIdx, err := frost.AggregateWithTweak(
			pegout.signingHashes[i],
			commitmentsPackages,
			hashOnlyShares,
			frost.NewPackage(pubkeyPackage),
			tapTweak,
		)
		if err != nil {
			if culpritFrostIdx != nil {
				culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
				s.logError(fmt.Sprintf("AggregateWithTweak failed. Culprit validator found: %d", culpritIdx), err)
				s.executeClaim(pegout, validatorIdx, culpritIdx)
			} else {
				s.logAggregateSignSharesError(err)
			}
			return false
		}

		signatures = append(signatures, signature)
	}

	if _, err := s.coordinator.SendSignatures(
		pegout.ID,
		validatorIdx,
		signatures,
	); err != nil {
		s.logSignatureSendError(err)
	} else {
		s.logSignatureSent(pegout.ID)
	}

	return false
}

func (s *SignService) Sign(
	publicKey []byte,
	tapTweak []byte,
	signingHash []byte,
	commitments map[uint16][]byte,
	nonceName string,
) ([]byte, *frost.Identifier, error) {
	secretPackage := s.keyStore.LoadSecret(publicKey)
	if secretPackage == nil {
		return nil, nil, fmt.Errorf("failed to load secret package by key %X", publicKey)
	}
	nonces := s.keyStore.LoadNonce(nonceName)
	if nonces == nil {
		return nil, nil, fmt.Errorf("failed to load nonce by name %s", nonceName)
	}
	return frost.SignWithTweak(
		frost.NewPackage(secretPackage),
		signingHash,
		helpers.ConvertMapToFrostPackages(commitments),
		frost.NewPackage(nonces),
		tapTweak,
	)
}

func (s *SignService) Commit(publicKey []byte) ([]byte, []byte, error) {
	secretPackage := s.keyStore.LoadSecret(publicKey)
	if secretPackage == nil {
		return nil, nil, fmt.Errorf("failed to load secret package by key %X", publicKey)
	}
	return frost.Commit(frost.NewPackage(secretPackage))
}

//
// Helpers
//

// Helper function to filter shares by all identifiers but
// for the given signing hash index
func filterSharesByHashIndex(
	// first key is the oracle identifier,
	// second key is the signing hash index
	shares map[uint16]map[uint16][]byte,
	index uint16,
) map[frost.Identifier]frost.Package {
	sharesMap := make(map[frost.Identifier]frost.Package)
	for identifier, participantShares := range shares {
		sharesMap[helpers.ValidatorIdxToFrost(identifier)] = frost.NewPackage(participantShares[index])
	}
	return sharesMap
}

// Helper function to get latest masterchain block
func (s *SignService) getLatestBlock(ctx context.Context) *ton.BlockIDExt {
	block, _ := s.ton.API.CurrentMasterchainInfo(ctx)
	return block
}

func (s *SignService) executeClaim(pegout *CachedPegout, validatorIdx uint16, culpritIdx uint16) {
	s.logExecuteClaim(pegout.ID)

	if s.ClaimCompleted(pegout, validatorIdx) {
		s.logMessage("claim completed")
		return
	}

	// claim package is not sent yet, send it to coordinator
	s.logSendClaim(pegout.ID, culpritIdx)

	if _, err := s.coordinator.SendSigningClaim(
		pegout.ID,
		validatorIdx,
		culpritIdx,
	); err != nil {
		s.logSigningClaimSentError(pegout.ID, err)
	} else {
		s.logSigningClaimSent(pegout.ID)
	}
}

func (s *SignService) executeResetPegoutSigning(pegoutID uint64, validatorIdx uint16) {
	s.logSendResetPegoutSigning(pegoutID)

	if _, err := s.coordinator.SendResetPegoutSigning(
		pegoutID,
		validatorIdx,
	); err != nil {
		s.logResetPegoutSigningSentError(pegoutID, err)
	} else {
		s.logResetPegoutSigningSent(pegoutID)
	}
}

func (s *SignService) ClaimCompleted(pegout *CachedPegout, validatorIdx uint16) bool {
	return pegout.artifacts.ClaimsMask.Bit(int(validatorIdx)) > 0
}
