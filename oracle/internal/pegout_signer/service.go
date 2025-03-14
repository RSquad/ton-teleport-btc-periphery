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
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
	"github.com/xssnick/tonutils-go/ton"
)

type SignService struct {
	dkgClient   *dkg.Client
	keyStore    keystore.Keystore
	coordinator *coordinator.CoordinatorContract
	validator   *validator.Validator
	ton         *tonclient.TonClient
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

	// Get first pegout record
	unsignedPegout := pegoutRecords[0]
	s.logProcessingPegout(&unsignedPegout)

	pubkeyPackage := dkg.R3.Data.PubkeyPackage

	validatorKeyInfo, err := s.validator.FindKeyInfo(dkg.VSet)
	if err != nil {
		s.logError("failed to get validator key", err)
		return
	}

	if validatorKeyInfo == nil {
		s.logOracleNotValidator(unsignedPegout.ID)
		return
	}

	s.coordinator.ConnectSigner(s.validator.GetSigner(validatorKeyInfo.KeyID))

	minSigners, err := helpers.CalcMinSigners(dkg.MaxSigners)
	if err != nil {
		s.logError("failed to calculate min signers", err)
		return
	}

	// Execute signing steps
	if s.doCommit(validatorKeyInfo, &unsignedPegout, minSigners) {
		if s.doSign(ctx, validatorKeyInfo, &unsignedPegout, minSigners) {
			if s.doAggregate(ctx, validatorKeyInfo, &unsignedPegout, pubkeyPackage) {
				s.logPegoutSigned(unsignedPegout.ID)
			}
		}
	}
}

func (s *SignService) doCommit(
	validatorKey *validator.KeyInfo,
	pegoutRecord *coordinator.PegoutRecord,
	minSigners uint16,
) bool {
	s.logCommitPegout(pegoutRecord.ID)
	identifier := validatorKey.PublicKey

	pegoutName := pegoutRecord.PegoutAddress.String()

	if pegoutRecord.HasCommitment(identifier) {
		s.logMessage("Commitment already exists")
		if pegoutRecord.CommitmentsCount() >= minSigners {
			s.logMessage("Commitment round completed")
			return true
		}
		return false
	}

	nonce := s.keyStore.LoadNonce(pegoutName)
	commitments := s.keyStore.LoadCommitments(pegoutName)

	if nonce == nil || commitments == nil {
		// If both nonce & commitments are not found in keystore
		if nonce == nil && commitments == nil {
			nonce, commitments, err := s.dkgClient.Commit(pegoutRecord.InternalKey)
			if err != nil {
				s.logError("commit call return error", err)
				return false
			}

			s.keyStore.StoreNonce(pegoutName, nonce)
			s.keyStore.StoreCommitments(pegoutName, commitments)
		} else {
			s.logErrNullNonceOrCommitments(nonce, commitments, pegoutName)
			return false
		}
	}

	if _, err := s.coordinator.SendCommitments(
		pegoutRecord.ID,
		validatorKey.VsetIdx,
		identifier,
		commitments,
	); err != nil {
		s.logError("failed to send commitments", err)
	} else {
		s.logCommitSent(pegoutRecord.ID)
	}

	return false
}

func (s *SignService) doSign(
	ctx context.Context,
	validatorKey *validator.KeyInfo,
	pegoutRecord *coordinator.PegoutRecord,
	minSigners uint16,
) bool {
	s.logMsgf("Sign pegout %x", pegoutRecord.ID)
	identifier := validatorKey.PublicKey
	pegoutName := pegoutRecord.PegoutAddress.String()

	if !pegoutRecord.HasCommitment(identifier) {
		s.logErrNoOracleCommitments(pegoutRecord.ID)
		return false
	}

	if pegoutRecord.SigningSharesCount() >= minSigners {
		s.logMinimalSharesReached(pegoutRecord.ID)
		return true
	}

	if pegoutRecord.HasSigningShare(identifier) {
		return false
	}

	pegoutContract := pegoutcontract.New(pegoutRecord.PegoutAddress, s.ton, ctx)
	block := s.getLatestBlock(ctx)
	signHashes, err := pegoutContract.GetSigningHashes(block)
	if err != nil {
		s.logError("failed to get signing hashes", err)
		return false
	}

	if len(signHashes) == 0 {
		s.logError("", fmt.Errorf("(pegoutId %x) no signing hashes", pegoutRecord.ID))
		return false
	}

	signShares := s.keyStore.LoadSigningShares(pegoutName)
	if signShares == nil {
		// Share for each signing hash is not generated yet.
		// Call frost.Sign for each signing hash

		txParts, err := pegoutContract.GetTxParts(block)
		if err != nil {
			s.logError("failed to get tx parts", err)
			return false
		}
		// Public key (X coord) used to get secret package from keystore
		publicKey := txParts.InternalKey
		// array with sign shares for each sign hash
		signShares = make([][]byte, 0, len(signHashes))
		i := 0
		// call frost.sign for each signing hash (or for each input)
		for _, input := range *txParts.Inputs {
			// tap tweak to tweak secret package before signing.
			tapTweak := input.BitcoinMerkleRoot
			// message to sign.
			message := signHashes[i]
			// nonce name used to load nonce from keystore
			nonceName := pegoutName
			signShare, err := s.dkgClient.Sign(publicKey, message, nonceName, pegoutRecord.Commitments, tapTweak)
			if err != nil {
				s.logError(fmt.Sprintf("failed to sign hash %d", i), err)
				return false
			}
			signShares = append(signShares, signShare)
			i = i + 1
		}
		s.keyStore.StoreSigningShares(pegoutName, signShares)
	}

	if _, err := s.coordinator.SendSigningShare(
		pegoutRecord.ID,
		validatorKey.VsetIdx,
		identifier,
		signShares,
	); err != nil {
		s.logError("failed to send signing share", err)
		return false
	}

	s.logSigningShareSent(pegoutRecord.ID)
	return false
}

func (s *SignService) doAggregate(
	ctx context.Context,
	validatorKey *validator.KeyInfo,
	pegoutRecord *coordinator.PegoutRecord,
	pubkeyPackage []byte,
) bool {
	s.logAggregateSignShares(pegoutRecord.ID)

	identifier := validatorKey.PublicKey
	pegoutContract := pegoutcontract.New(pegoutRecord.PegoutAddress, s.ton, ctx)
	block := s.getLatestBlock(ctx)
	txParts, err := pegoutContract.GetTxParts(block)
	if err != nil {
		s.logError("failed to get tx parts", err)
		return false
	}

	if len(*txParts.Signatures) > 0 {
		return true
	}

	signHashes, err := pegoutContract.GetSigningHashes(block)

	commitmentsMap := helpers.ConvertMapToFrostPackages(pegoutRecord.Commitments)

	signatures := make([][]byte, 0, len(signHashes))

	for i, hash := range signHashes {
		frostShares := filterSharesByIndex(pegoutRecord.SigningShares, uint8(i))
		// TODO tapTweak := (*txParts.Inputs)[i].BitcoinMerkleRoot
		signature, err := frost.AggregateWithTweak(hash, commitmentsMap, frostShares, frost.NewPackage(pubkeyPackage))
		if err != nil {
			s.logAggregateSignSharesError(err)
			return false
		}
		signatures = append(signatures, signature)
	}

	if _, err := s.coordinator.SendSignatures(
		pegoutRecord.ID,
		validatorKey.VsetIdx,
		identifier,
		signatures,
	); err != nil {
		s.logSignatureSendError(err)
	} else {
		s.logSignatureSent(pegoutRecord.ID)
	}

	return false
}

//
// Helpers
//

// Helper function to filter shares by all identifiers but
// for the given signing hash index
func filterSharesByIndex(
	// first key is the oracle identifier,
	// second key is the signing hash index
	shares map[string]map[uint8][]byte,
	signingHashIndex uint8,
) map[frost.Identifier]frost.Package {
	sharesMap := make(map[frost.Identifier]frost.Package)
	for identifier, shares := range shares {
		frostIdentifier, _ := frost.DecodeIdentifier(identifier)
		sharesMap[*frostIdentifier] = frost.NewPackage(shares[signingHashIndex])
	}
	return sharesMap
}

// Helper function to get latest masterchain block
func (s *SignService) getLatestBlock(ctx context.Context) *ton.BlockIDExt {
	block, _ := s.ton.API.CurrentMasterchainInfo(ctx)
	return block
}
