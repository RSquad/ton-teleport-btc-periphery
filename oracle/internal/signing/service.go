package signing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
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
	inputs        []pegoutcontract.TxPartsInput
	tx            *pegoutcontract.TxParts
	signingHashes [][]byte
	artifacts     *coordinator.PegoutRecord
}

type SignService struct {
	keyStore     keystore.Keystore
	coordinator  *coordinator.CoordinatorContract
	validator    *validator.Validator
	ton          *tonclient.TonClient
	cachedPegout *CachedPegout
}

func NewService(
	keyStore keystore.Keystore,
	validator *validator.Validator,
	coordinator *coordinator.CoordinatorContract,
	tonclient *tonclient.TonClient,
) *SignService {
	return &SignService{
		keyStore:    keyStore,
		validator:   validator,
		coordinator: coordinator,
		ton:         tonclient,
	}
}

func (s *SignService) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer logger.DefaultLogFinishWork("SignService")
	logger.DefaultLogStartWork("SignService")

	// A periodic event that triggers every 6 seconds to call the ExecuteSign() function
	ticker := time.NewTicker(6 * time.Second)
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
	txInputs, err := pegoutContract.GetInputs(block)
	if err != nil {
		s.logError("failed to cache pegout", err)
		return nil, err
	}
	signingHashes, err := pegoutContract.GetSigningHashes(block)
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

	if len(pegoutRecords) == 0 {
		s.logMessage("No sign requests")
		return
	}

	s.validator.GetSessionSigner().OnNewDKG(dkg.Until.Unix())
	s.logSigningRequestsCount(len(pegoutRecords))

	// Get oldest unsigned pegout record
	unsignedPegout := pegoutRecords[0]
	s.logProcessingPegout(&unsignedPegout)
	cachedPegout, err := s.cachePegout(ctx, &unsignedPegout)
	if err != nil {
		s.logError("failed to cache pegout", err)
		return
	}
	if cachedPegout == nil {
		panic("cached pegout is nil")
	}

	pubkeyPackage := dkg.R3.Data.PubkeyPackage

	validatorKeyInfo, err := s.validator.FindKeyInfo(dkg.VSet)
	if err != nil {
		s.logError("failed to get validator key", err)
		return
	}

	if validatorKeyInfo == nil {
		s.logOracleNotValidator(cachedPegout.ID)
		return
	}

	s.coordinator.ConnectSigner(s.validator.GetSessionSigner())

	minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
	if err != nil {
		s.logError("failed to calculate min signers", err)
		return
	}

	// Execute signing steps
	if s.doCommit(validatorKeyInfo, cachedPegout, minSigners) {
		if s.doSign(ctx, validatorKeyInfo, cachedPegout, minSigners) {
			if s.doAggregate(ctx, validatorKeyInfo, cachedPegout, pubkeyPackage) {
				s.logPegoutSigned(cachedPegout.ID)
			}
		}
	}
}

func (s *SignService) doCommit(
	validatorKey *validator.KeyInfo,
	pegout *CachedPegout,
	minSigners uint16,
) bool {
	s.logCommitPegout(pegout.ID)

	localIdentifier := helpers.ValidatorIdxToFrost(validatorKey.VsetIdx)

	if pegout.artifacts.HasCommitment(localIdentifier) {
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
		validatorKey.VsetIdx,
		localIdentifier,
		commitments,
	); err != nil {
		s.logError("failed to send commitments", err)
	} else {
		s.logCommitSent(pegout.ID)
	}

	return false
}

func (s *SignService) doSign(
	ctx context.Context,
	validatorKey *validator.KeyInfo,
	pegout *CachedPegout,
	minSigners uint16,
) bool {
	s.logSignPegout(pegout.ID)

	localIdentifier := helpers.ValidatorIdxToFrost(validatorKey.VsetIdx)

	if !pegout.artifacts.HasCommitment(localIdentifier) {
		s.logErrNoOracleCommitments(pegout.ID)
		return false
	}

	if pegout.artifacts.SigningSharesCount() >= int(minSigners) {
		s.logMinimalSharesReached(pegout.ID)
		return true
	}

	if pegout.artifacts.HasSigningShare(localIdentifier) {
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
			signShare, err := s.Sign(
				publicKey,                    // public key
				input.BitcoinMerkleRoot,      // tap merkle root used as tweak for tweaking signing share
				pegout.signingHashes[i],      // signing hash for current input
				pegout.artifacts.Commitments, // oracle's commitments to sign pegout hashes
				pegout.addrStr,               // pegout address used as key to load nonce from keystore
			)
			if err != nil {
				s.logError(fmt.Sprintf("failed to sign hash %d", i), err)
				return false
			}
			signShares = append(signShares, signShare)
		}
		s.keyStore.StoreSigningShares(pegout.addrStr, signShares)
	}

	s.logSendSigningShare(pegout.ID, signShares)
	if _, err := s.coordinator.SendSigningShare(
		pegout.ID,
		validatorKey.VsetIdx,
		localIdentifier,
		signShares,
	); err != nil {
		s.logSendSigningShareError(pegout.ID, err)
	} else {
		s.logSigningShareSent(pegout.ID)
	}

	return false
}

func (s *SignService) doAggregate(
	ctx context.Context,
	validatorKey *validator.KeyInfo,
	pegout *CachedPegout,
	pubkeyPackage []byte,
) bool {
	s.logAggregateSignShares(pegout.ID)

	commitmentsPackages := helpers.ConvertMapToFrostPackages(pegout.artifacts.Commitments)
	signatures := make([][]byte, 0, len(pegout.signingHashes))

	for i, input := range pegout.inputs {
		hashOnlyShares := filterSharesByHashIndex(pegout.artifacts.SigningShares, i)
		tapTweak := input.BitcoinMerkleRoot
		signature, err := frost.AggregateWithTweak(
			pegout.signingHashes[i],
			commitmentsPackages,
			hashOnlyShares,
			frost.NewPackage(pubkeyPackage),
			tapTweak,
		)
		if err != nil {
			s.logAggregateSignSharesError(err)
			return false
		}
		signatures = append(signatures, signature)
	}

	if _, err := s.coordinator.SendSignatures(
		pegout.ID,
		validatorKey.VsetIdx,
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
	commitments map[string][]byte,
	nonceName string,
) ([]byte, error) {
	secretPackage := s.keyStore.LoadSecret(publicKey)
	if secretPackage == nil {
		return nil, fmt.Errorf("failed to load secret package by key %x", publicKey)
	}
	nonces := s.keyStore.LoadNonce(nonceName)
	if nonces == nil {
		return nil, fmt.Errorf("failed to load nonce by name %s", nonceName)
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
		return nil, nil, fmt.Errorf("failed to load secret package by key %x", publicKey)
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
	shares map[string]map[int][]byte,
	index int,
) map[frost.Identifier]frost.Package {
	sharesMap := make(map[frost.Identifier]frost.Package)
	for identifier, participantShares := range shares {
		frostIdentifier, _ := frost.DecodeIdentifier(identifier)
		sharesMap[*frostIdentifier] = frost.NewPackage(participantShares[index])
	}
	return sharesMap
}

// Helper function to get latest masterchain block
func (s *SignService) getLatestBlock(ctx context.Context) *ton.BlockIDExt {
	block, _ := s.ton.API.CurrentMasterchainInfo(ctx)
	return block
}
