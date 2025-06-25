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
	nonces        [][]byte
	commitments   [][]byte
}

type SignService struct {
	keyStore          keystore.Keystore
	coordinator       coordinator.Coordinator
	ton               *tonclient.TonClient
	cachedPegout      *CachedPegout
	executeSignPeriod int64 // `period` in seconds to call the ExecuteSign() function
	dkgUntil          time.Time
	sessionSigner     *validator.SessionSigner
}

func NewService(
	keyStore keystore.Keystore,
	coordinator coordinator.Coordinator,
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

func (s *SignService) cleanupNonces() {
	if s.cachedPegout == nil || s.cachedPegout.nonces == nil {
		return
	}

	for i := range s.cachedPegout.nonces {
		if s.cachedPegout.nonces[i] != nil {
			for j := range s.cachedPegout.nonces[i] {
				s.cachedPegout.nonces[i][j] = 0
			}
		}
	}
	s.cachedPegout.nonces = nil
}

func (s *SignService) cleanupCommitments() {
	if s.cachedPegout == nil || s.cachedPegout.commitments == nil {
		return
	}
	s.cachedPegout.commitments = nil
}

func (s *SignService) cachePegout(
	ctx context.Context,
	unsignedPegout *coordinator.PegoutRecord,
) (*CachedPegout, error) {
	if s.cachedPegout != nil && s.cachedPegout.ID == unsignedPegout.ID {
		// If the pegout expired, cleanup the nonces, commitments and signing shares
		if (s.cachedPegout.artifacts.ExpiredAt != time.Unix(0, 0)) && !s.cachedPegout.artifacts.ExpiredAt.Equal(unsignedPegout.ExpiredAt) {
			s.cleanupNonces()
			s.cleanupCommitments()
			s.keyStore.Cleanup()
		}
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
	s.coordinator.ConnectSigner(s.sessionSigner)

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

	minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
	if err != nil {
		s.logError("failed to calculate min signers", err)
		return
	}

	// Execute signing steps
	s.logDebug("Try running the Pegout Signing rounds")

	if s.doCommit(validatorIdx, minSigners) {
		if s.doSign(validatorIdx, minSigners) {
			s.doAggregate(validatorIdx, pubkeyPackage)
		}
	}
}

func (s *SignService) doCommit(
	validatorIdx uint16,
	minSigners uint16,
) bool {
	s.logDebug("Pegout Signing: Commit")
	pegout := s.cachedPegout
	s.logCommitPegout(pegout.ID)

	if pegout.artifacts.HasCommitment(validatorIdx) {
		s.logMessage("Commitment already exists")

		if pegout.artifacts.CommitmentsCount() >= minSigners {
			s.logMinimalCommitmentsReached(pegout, minSigners)
			return true
		} else {
			s.logMinimalCommitmentsWaitingForOtherOracles(pegout, minSigners)
			return false
		}
	}

	err := s.generateCommitments()
	if err != nil {
		s.logError("failed to generate commitments", err)
		return false
	}

	s.sendCommitments(pegout, validatorIdx)

	return false
}

func (s *SignService) doSign(
	validatorIdx uint16,
	minSigners uint16,
) bool {
	s.logDebug("Pegout Signing: Sign")
	pegout := s.cachedPegout
	s.logSignPegout(pegout.ID)

	if !pegout.artifacts.HasCommitment(validatorIdx) {
		s.logErrNoOracleCommitments(pegout.ID)
		return false
	}

	if pegout.artifacts.HasSigningShare(validatorIdx) {
		s.logSigningShareAlreadyExists(pegout.ID)

		if pegout.artifacts.SigningSharesCount() >= int(minSigners) {
			s.logMinimalSharesReached(pegout, minSigners)
			return true
		} else {
			s.logMinimalSharesWaitingForOtherOracles(pegout, minSigners)
			return false
		}
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

		// array with sign shares for each signing hash
		signShares = make([][]byte, 0, len(pegout.signingHashes))
		// generate signing share for each input
		for i := range pegout.inputs {
			// generate signing share for the i-th input
			signShare, err := s.SignInput(validatorIdx, i)
			if err != nil {
				s.logSignError(i, err)
				return false
			}
			signShares = append(signShares, signShare)
		}
		s.keyStore.StoreSigningShares(pegout.addrStr, signShares)
	}

	s.sendSigningShares(pegout, validatorIdx, signShares)

	return false
}

func (s *SignService) doAggregate(
	validatorIdx uint16,
	pubkeyPackage []byte,
) {
	s.logDebug("Pegout Signing: Aggregate")
	pegout := s.cachedPegout
	s.logAggregateSignShares()
	s.cleanupNonces()

	// Check if the signatures have already been sent
	if checkSignaturesMask(pegout, validatorIdx) {
		s.logSignaturesSent(pegout.ID, pegout.artifacts.Signatures.Count, pegout.artifacts.MaxSigners)
		return
	}

	signatures := make([][]byte, 0, len(pegout.signingHashes))

	for i := range pegout.inputs {
		signature, err := s.aggregateSignatureForInput(i, validatorIdx, pubkeyPackage)
		if err != nil {
			s.logAggregateSignSharesError(i, err)
			return
		}
		signatures = append(signatures, signature)
	}

	// Send signatures
	s.SendSignatures(pegout, validatorIdx, signatures)
}

func (s *SignService) sendCommitments(pegout *CachedPegout, validatorIdx uint16) {
	packedCommitments, err := helpers.SerializeCommitments(pegout.commitments, helpers.FrostCommitmentLength)
	if err != nil {
		s.logError("failed to serialize commitments", err)
		return
	}
	s.logSendCommitments(pegout.ID)
	if _, err := s.coordinator.SendCommitments(
		pegout.ID,
		pegout.artifacts.ExpiredAt.Unix(),
		validatorIdx,
		packedCommitments,
	); err != nil {
		s.logSendCommitmentsError(pegout.ID, err)
	} else {
		s.logCommitSent(pegout.ID)
	}
}

func (s *SignService) sendSigningShares(pegout *CachedPegout, validatorIdx uint16, signShares [][]byte) {
	s.logSendSigningShare(pegout.ID, signShares)
	if _, err := s.coordinator.SendSigningShare(
		pegout.ID,
		pegout.artifacts.ExpiredAt.Unix(),
		validatorIdx,
		signShares,
	); err != nil {
		s.logSendSigningShareError(pegout.ID, err)
	} else {
		s.logSigningShareSent(pegout.ID)
	}
}

func (s *SignService) SendSignatures(pegout *CachedPegout, validatorIdx uint16, signatures [][]byte) int {
	_, err := s.coordinator.SendSignatures(
		pegout.ID,
		pegout.artifacts.ExpiredAt.Unix(),
		validatorIdx,
		signatures,
	)
	if err != nil {
		exitCode, _ := helpers.ExtractExitCode(err.Error())
		if exitCode == helpers.DifferentPegoutSignatures {
			s.sendClaimBySignatureMask(pegout, validatorIdx)
		} else {
			s.logSignatureSendError(pegout.ID, err)
		}

		return exitCode
	}

	s.logSignatureSent(pegout.ID)
	return 0
}

func (s *SignService) generateCommitments() error {
	pegout := s.cachedPegout
	if pegout.nonces == nil || pegout.commitments == nil {
		// If both nonce & commitments are not found in keystore
		if pegout.nonces == nil && pegout.commitments == nil {
			s.logMessage("generate commitments")
			nonces, commitments, err := s.Commit(pegout.tx.InternalKey, len(pegout.inputs))
			if err != nil {
				return fmt.Errorf("commit error: %w", err)
			}
			pegout.nonces = nonces
			pegout.commitments = commitments
		} else {
			if pegout.nonces == nil {
				return fmt.Errorf("failed to load nonce for %s", pegout.addrStr)
			} else if pegout.commitments == nil {
				return fmt.Errorf("failed to load commitments for %s", pegout.addrStr)
			}
		}
	}
	return nil
}

func (s *SignService) SignInput(validatorIdx uint16, inputIndex int) ([]byte, error) {
	pegout := s.cachedPegout
	// Public key (X coord) used to get secret package from keystore
	publicKey := s.cachedPegout.tx.InternalKey

	// get commitments from all validators for the inputIndex
	inputCommitments, err := helpers.DeserializeInputCommitmentForAll(pegout.artifacts.Commitments, len(pegout.inputs), inputIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize input commitments: %w", err)
	}

	// Validate inputIndex
	pegoutSigningHashesLen := len(pegout.signingHashes)
	if pegoutSigningHashesLen <= inputIndex {
		return nil, fmt.Errorf("len(pegout.signingHashes)(%d) <= inputIndex(%d)", pegoutSigningHashesLen, inputIndex)
	}

	pegoutNoncesLen := len(pegout.nonces)
	if pegoutNoncesLen <= inputIndex {
		return nil, fmt.Errorf("len(pegout.nonces)(%d) <= inputIndex(%d)", pegoutNoncesLen, inputIndex)
	}

	pegoutInputsLen := len(pegout.inputs)
	if pegoutInputsLen <= inputIndex {
		return nil, fmt.Errorf("len(pegout.inputs)(%d) <= inputIndex(%d)", pegoutInputsLen, inputIndex)
	}

	secretPackage := s.keyStore.LoadSecret(publicKey)
	if secretPackage == nil {
		return nil, fmt.Errorf("failed to load secret package by key %X", publicKey)
	}
	share, culpritFrostIdx, err := frost.SignWithTweak(
		frost.NewPackage(secretPackage),
		pegout.signingHashes[inputIndex],
		helpers.ConvertMapToFrostPackages(inputCommitments),
		frost.NewPackage(pegout.nonces[inputIndex]),
		pegout.inputs[inputIndex].Data.BitcoinMerkleRoot,
	)
	if err != nil {
		if culpritFrostIdx != nil {
			culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
			s.logError(fmt.Sprintf("Culprit validator found: %d", culpritIdx), err)
			s.executeClaim(pegout, validatorIdx, culpritIdx)
		}
		return nil, err
	}
	return share, nil
}

func (s *SignService) Commit(publicKey []byte, inputsCount int) ([][]byte, [][]byte, error) {
	secretPackage := s.keyStore.LoadSecret(publicKey)
	if secretPackage == nil {
		return nil, nil, fmt.Errorf("failed to load secret package by key %X", publicKey)
	}
	nonces := make([][]byte, inputsCount)
	commitments := make([][]byte, inputsCount)
	var err error
	frostPackage := frost.NewPackage(secretPackage)
	for i := 0; i < inputsCount; i++ {
		nonces[i], commitments[i], err = frost.Commit(frostPackage)
		if err != nil {
			return nil, nil, err
		}
	}
	return nonces, commitments, nil
}

func (s *SignService) aggregateSignatureForInput(inputIndex int, validatorIdx uint16, pubkeyPackage []byte) ([]byte, error) {
	pegout := s.cachedPegout
	inputCommitments, err := helpers.DeserializeInputCommitmentForAll(pegout.artifacts.Commitments, len(pegout.inputs), inputIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize input commitments for input %d: %w", inputIndex, err)
	}
	commitmentsPackages := helpers.ConvertMapToFrostPackages(inputCommitments)
	hashOnlyShares := filterSharesByHashIndex(pegout.artifacts.SigningShares, uint16(inputIndex))
	tapTweak := pegout.inputs[inputIndex].Data.BitcoinMerkleRoot
	signature, culpritFrostIdx, err := frost.AggregateWithTweak(
		pegout.signingHashes[inputIndex],
		commitmentsPackages,
		hashOnlyShares,
		frost.NewPackage(pubkeyPackage),
		tapTweak,
	)
	if err != nil {
		if culpritFrostIdx != nil {
			culpritIdx := helpers.FrostToValidatorIdx(*culpritFrostIdx)
			s.logError(fmt.Sprintf("Culprit validator found: %d", culpritIdx), err)
			s.executeClaim(pegout, validatorIdx, culpritIdx)
		}
		return nil, err
	}
	return signature, nil
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
		s.logMessage("claim completed (it has already been sent)")
		return
	}

	// claim package is not sent yet, send it to coordinator
	s.logSendClaim(pegout.ID, culpritIdx)

	if _, err := s.coordinator.SendSigningClaim(
		pegout.ID,
		pegout.artifacts.ExpiredAt.Unix(),
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

func (s *SignService) sendClaimBySignatureMask(pegout *CachedPegout, validatorIdx uint16) {
	// Culprit Oracle id = index of the first non zero bit in pegout.artifacts.Signatures.Mask
	culpritIdx := pegout.artifacts.Signatures.Mask.BitLen() - 1
	s.logError(fmt.Sprintf("Signature sending failed. Culprit validator identified: %d. The signature sent by the culprit validator differs from the calculated signature", culpritIdx), nil)

	// Sent claim
	s.executeClaim(pegout, validatorIdx, uint16(culpritIdx))
}

func checkSignaturesMask(pegout *CachedPegout, validatorIdx uint16) bool {
	return pegout.artifacts.Signatures.Mask.Bit(int(validatorIdx)) == 1
}
