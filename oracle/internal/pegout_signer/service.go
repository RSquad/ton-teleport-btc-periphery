package pegoutsigner

import (
	"context"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
	"github.com/xssnick/tonutils-go/ton"
)

type CachedPegout struct {
	ID            uint64
	name          string
	inputs        []pegoutcontract.TxPartsInput
	tx            *pegoutcontract.TxParts
	signingHashes [][]byte
	artifacts     *coordinator.PegoutRecord
}

type SignService struct {
	dkgClient    *dkg.Client
	keyStore     keystore.Keystore
	coordinator  *coordinator.CoordinatorContract
	validator    *validator.Validator
	ton          *tonclient.TonClient
	cachedPegout *CachedPegout
}

func NewService(
	dkgClient *dkg.Client,
	keyStore keystore.Keystore,
	validator *validator.Validator,
	coordinator *coordinator.CoordinatorContract,
	tonclient *tonclient.TonClient,
) *SignService {
	return &SignService{
		dkgClient:   dkgClient,
		keyStore:    keyStore,
		validator:   validator,
		coordinator: coordinator,
		ton:         tonclient,
	}
}

func (s *SignService) Work(ctx context.Context) {
	logger.DefaultLogStartWork("SignService")
	defer logger.DefaultLogFinishWork("SignService", nil)

	for {
		s.ExecuteSign(ctx)
		time.Sleep(6 * time.Second)
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

	name := unsignedPegout.PegoutAddress.String()

	s.cachedPegout = &CachedPegout{
		ID:            unsignedPegout.ID,
		name:          name,
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

	validatorKeyInfo, err := s.validator.FindKeyInfo(dkg.VSet, &s.keyStore)
	if err != nil {
		s.logError("failed to get validator key", err)
		return
	}

	if validatorKeyInfo == nil {
		s.logOracleNotValidator(cachedPegout.ID)
		return
	}

	// TODO: use Oracle sign
	s.coordinator.ConnectSigner(s.validator.GetSigner(validatorKeyInfo.KeyID, &s.keyStore, validator.SIGNER_ORACLE))

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
	identifier := validatorKey.PublicKey

	if pegout.artifacts.HasCommitment(identifier) {
		s.logMessage("Commitment already exists")
		if pegout.artifacts.CommitmentsCount() >= minSigners {
			s.logMessage("Commitment round completed")
			return true
		}
		s.logMessage("Waiting for other oracles to commit")
		return false
	}

	nonce := s.keyStore.LoadNonce(pegout.name)
	commitments := s.keyStore.LoadCommitments(pegout.name)

	if nonce == nil || commitments == nil {
		// If both nonce & commitments are not found in keystore
		if nonce == nil && commitments == nil {
			s.logMessage("generate commitments")
			var err error
			nonce, commitments, err = s.dkgClient.Commit(pegout.tx.InternalKey)
			if err != nil {
				s.logError("commit call return error", err)
				return false
			}
			s.keyStore.StoreNonce(pegout.name, nonce)
			s.keyStore.StoreCommitments(pegout.name, commitments)
		} else {
			s.logErrNullNonceOrCommitments(nonce, commitments, pegout.name)
			return false
		}
	}

	s.logSendCommitments(pegout.ID, commitments)
	if _, err := s.coordinator.SendCommitments(
		pegout.ID,
		validatorKey.VsetIdx,
		identifier,
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
	identifier := validatorKey.PublicKey

	if !pegout.artifacts.HasCommitment(identifier) {
		s.logErrNoOracleCommitments(pegout.ID)
		return false
	}

	if pegout.artifacts.SigningSharesCount() >= int(minSigners) {
		s.logMinimalSharesReached(pegout.ID)
		return true
	}

	if pegout.artifacts.HasSigningShare(identifier) {
		s.logSigningShareAlreadyExists(pegout.ID)
		return false
	}

	if len(pegout.signingHashes) == 0 {
		s.logErrNothingToSign(pegout.ID)
		return false
	}

	signShares := s.keyStore.LoadSigningShares(pegout.name)
	if signShares == nil {
		s.logMessage("generate signing share")
		// Share for each signing hash is not generated yet.
		// Call frost.Sign for each signing hash

		// Public key (X coord) used to get secret package from keystore
		publicKey := pegout.tx.InternalKey
		// array with sign shares for each sign hash
		signShares = make([][]byte, 0, len(pegout.signingHashes))
		// call frost.sign for each signing hash (or for each input)
		for i, input := range pegout.inputs {
			// tap tweak to tweak secret package before signing.
			tapTweak := input.BitcoinMerkleRoot
			// message to sign.
			message := pegout.signingHashes[i]
			// nonce name used to load nonce
			nonceName := pegout.name
			signShare, err := s.dkgClient.Sign(publicKey, message, nonceName, pegout.artifacts.Commitments, tapTweak)
			if err != nil {
				s.logError(fmt.Sprintf("failed to sign hash %d", i), err)
				return false
			}
			signShares = append(signShares, signShare)
		}
		s.keyStore.StoreSigningShares(pegout.name, signShares)
	}

	s.logSendSigningShare(pegout.ID, signShares)
	if _, err := s.coordinator.SendSigningShare(
		pegout.ID,
		validatorKey.VsetIdx,
		identifier,
		signShares,
	); err != nil {
		s.logError("failed to send signing share", err)
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
