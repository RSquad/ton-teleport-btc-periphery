package pegoutsigner

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
	"github.com/xssnick/tonutils-go/ton"
)

type SignService struct {
	mu          sync.Mutex
	dkgClient   *dkg.Client
	config      *cfg.Cfg
	keyStore    keystore.Keystore
	coordinator *coordinator.CoordinatorContract
	validator   *validator.Validator
	ton         *tonclient.TonClient
}

func NewService(
	config *cfg.Cfg,
	dkgClient *dkg.Client,
	keyStore keystore.Keystore,
	validator *validator.Validator,
	coordinator *coordinator.CoordinatorContract,
	tonclient *tonclient.TonClient,
) *SignService {
	return &SignService{
		config:      config,
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
		time.Sleep(3 * time.Second)
	}
}

func (s *SignService) ExecuteSign(ctx context.Context) {
	defer func() {
		s.logMessage("completed")
	}()

	s.logMessage("started")

	dkg, err := s.coordinator.GetPrevDKG()
	if err != nil {
		s.logMessage(fmt.Sprintf("failed to get previous DKG: %w", err))
		return
	}

	if dkg == nil {
		s.logMessage("DKG not yet completed")
		return
	}
	s.execute(ctx, dkg)
}

func (s *SignService) execute(ctx context.Context, dkg *coordinator.DKG) {
	pegoutRecords, err := s.coordinator.GetUnsignedPegouts()
	if err != nil {
		s.logMessage(fmt.Sprintf("failed to get unsigned pegouts: %w", err))
		return
	}

	if len(pegoutRecords) == 0 {
		s.logMessage("No sign requests")
		return
	}

	s.logMessage(fmt.Sprintf("%d signing requests", len(pegoutRecords)))

	// Get first pegout record
	unsignedPegout := pegoutRecords[0]
	s.logProcessingPegout(&unsignedPegout)

	valKey, err := s.validator.GetValidatorKey(dkg)
	if err != nil {
		s.logError("failed to get validator key", err)
		return
	}

	if valKey == nil {
		s.logOracleNotValidator(unsignedPegout.ID)
		return
	}

	s.coordinator.ConnectSigner(s.validator.GetSigner(valKey.ValidatorID))

	minSigners := int(math.Floor(float64(dkg.MaxSigners) * 2 / 3))

	// Execute signing steps
	if s.doCommit(valKey, &unsignedPegout, minSigners) {
		if s.doSign(ctx, valKey, &unsignedPegout, minSigners) {
			if s.doAggregate(ctx, valKey, &unsignedPegout) {
				s.logPegoutSigned(unsignedPegout.ID)
			}
		}
	}
}

func (s *SignService) doCommit(
	validatorKey *validator.ValidatorKey,
	pegoutRecord *coordinator.PegoutRecord,
	minSigners int,
) bool {
	s.logCommitPegout(pegoutRecord.ID)
	identifier := validatorKey.ValidatorKey

	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if pegoutRecord.HasCommitment(identifier) {
		s.logMessage("Commitment already exists")
		if pegoutRecord.CommitmentsCount() >= minSigners {
			s.logMessage("Commitment round completed")
			return true
		}
		return false
	}

	nonce := s.keyStore.LoadNonce(pegoutAddressStr)
	commitments := s.keyStore.LoadCommitments(pegoutAddressStr)

	if nonce == nil || commitments == nil {
		// If both nonce & commitments are not found in keystore
		if nonce == nil && commitments == nil {
			nonce, commitments, err := s.dkgClient.Commit(pegoutRecord.InternalKey)
			if err != nil {
				s.logError("commit call return error", err)
				return false
			}

			// TODO
			// s.keyStore.StoreNonce(pegoutAddressStr, nonce)
			// s.keyStore.StoreCommitments(pegoutAddressStr, commitments)
		} else {
			s.logErrNullNonceOrCommitments(nonce, commitments, pegoutAddressStr)
			return false
		}
	}

	if _, err := s.coordinator.SendCommitments(
		pegoutRecord.ID,
		validatorKey.ValidatorIdx,
		identifier,
		commitments,
	); err != nil {
		s.logError("failed to send commitments", err)
		return false
	}

	s.logMessage(fmt.Sprintf("Commit sent for pegout %x", pegoutRecord.ID))
	return false
}

func (s *SignService) doSign(
	ctx context.Context,
	validatorKey *validator.ValidatorKey,
	pegoutRecord *coordinator.PegoutRecord,
	minSigners int,
) bool {
	s.logMsgf("Sign pegout %x", pegoutRecord.ID)
	identifier := validatorKey.ValidatorKey
	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if !pegoutRecord.HasCommitment(identifier) {
		s.logErrNoOracleCommitments(pegoutRecord.ID)
		return false
	}

	if pegoutRecord.SigningSharesCount() >= minSigners {
		s.logMessage(fmt.Sprintf("Moving to aggregation phase for pegout %x", pegoutRecord.ID))
		return true
	}

	if !pegoutRecord.HasSigningShare(identifier) {
		return false
	}

	commitmentsPackages := make([]CommitmentPackage, 0)
	for key, commitment := range pegoutRecord.Commitments {
		commitmentsPackages = append(commitmentsPackages, CommitmentPackage{
			Identifier: key,
			Package:    commitment,
		})
	}

	pegoutContract := pegoutcontract.New(pegoutRecord.PegoutAddress, s.ton, ctx)
	block := s.getLatestBlock(ctx)
	signHashes, err := pegoutContract.GetSigningHashes(block)
	if err != nil {
		s.logError("failed to get signing hashes", err)
		return false
	}

	if len(signHashes) == 0 {
		s.logError("", fmt.Errorf("No signing hashes in pegout %x", pegoutRecord.ID))
		return false
	}

	signShares := s.keyStore.LoadSigningShares(pegoutAddressStr)
	if signShares == nil {
		txParts, err := pegoutContract.GetTxParts(block)
		if err != nil {
			s.logError("failed to get tx parts", err)
			return false
		}

		publicKey := txParts.InternalKey
		signShares = make([][]byte, 0, len(signHashes))
		i := 0
		for _, input := range *txParts.Inputs {
			tapTweak := input.BitcoinMerkleRoot
			signShare, err := s.dkgClient.Sign(publicKey, tapTweak, pegoutAddressStr, signShares[i])
			if err != nil {
				s.logError(fmt.Sprintf("failed to sign hash %d", i), err)
				return false
			}
			signShares = append(signShares, signShare)
			i = i + 1
		}
		s.keyStore.StoreSigningShares(pegoutAddressStr, signShares)
	}

	if _, err := s.coordinator.SendSigningShare(
		pegoutRecord.ID,
		validatorKey.ValidatorIdx,
		identifier,
		signShares,
	); err != nil {
		s.logError("failed to send signing share", err)
		return false
	}

	s.logMessage(fmt.Sprintf("Signing share sent for pegout %x", pegoutRecord.ID))
	return false
}

func (s *SignService) doAggregate(ctx context.Context, validatorKey *validator.ValidatorKey, pegoutRecord *coordinator.PegoutRecord) bool {
	s.logMessage(fmt.Sprintf("Aggregate sign shares for pegout %x", pegoutRecord.ID))
	identifier := validatorKey.ValidatorKey

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

	prevDkg, err := s.coordinator.GetPrevDKG()
	if err != nil {
		// TODO: log error
		return false
	}

	pubkeyPackage := prevDkg.R3.Data.PubkeyPackage

	commitmentsMap := make(map[frost.Identifier]frost.Package)
	for key, pkg := range pegoutRecord.Commitments {
		frostIdentifier, _ := frost.DecodeIdentifier(key)
		commitmentsMap[*frostIdentifier] = frost.NewPackage(pkg)
	}

	signatures := make([][]byte, 0, len(signHashes))
	i := 0
	for _, shares := range pegoutRecord.SigningShares {
		sharesMap := make(map[frost.Identifier]frost.Package)
		for key, pkg := range shares {
			frostIdentifier, _ := frost.DecodeIdentifier(key)
			sharesMap[*frostIdentifier] = frost.NewPackage(pkg)
		}

		hash := signHashes[i]
		// TODO tapTweak := (*txParts.Inputs)[i].BitcoinMerkleRoot
		signature, err := frost.AggregateWithTweak(hash, commitmentsMap, sharesMap, frost.NewPackage(pubkeyPackage))
		if err != nil {
			s.logError("failed to aggregate signatures", err)
			return false
		}
		signatures = append(signatures, signature)
		i = i + 1
	}

	if _, err := s.coordinator.SendSignatures(
		pegoutRecord.ID,
		validatorKey.ValidatorIdx,
		identifier,
		signatures,
	); err != nil {
		s.logError("failed to send signatures", err)
		return false
	}

	s.logMessage(fmt.Sprintf("Signature sent for pegout %x", pegoutRecord.ID))
	return true
}

// Helper function to filter shares by index
func filterSharesByIndex(shares []SigningShare, targetIndex int) []SigningShare {
	filtered := make([]SigningShare, 0)
	for _, share := range shares {
		if share.Index == fmt.Sprintf("%d", targetIndex) {
			filtered = append(filtered, share)
		}
	}
	return filtered
}

// Helper function to get latest masterchain block
func (s *SignService) getLatestBlock(ctx context.Context) *ton.BlockIDExt {
	block, _ := s.ton.API.CurrentMasterchainInfo(ctx)
	return block
}
