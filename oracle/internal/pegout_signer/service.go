package pegoutsigner

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator"
)

type SignService struct {
	mu          sync.Mutex
	dkgClient   *dkg.Client
	config      *cfg.Cfg
	keyStore    keystore.Keystore
	coordinator *coordinator.CoordinatorContract
	validator   *validator.Validator
	ctx         context.Context
}

func NewService(
	config *cfg.Cfg,
	dkgClient *dkg.Client,
	keyStore keystore.Keystore,
	validator *validator.Validator,
	coordinator *coordinator.CoordinatorContract,
	ctx context.Context,
) *SignService {
	return &SignService{
		config:      config,
		dkgClient:   dkgClient,
		keyStore:    keyStore,
		validator:   validator,
		coordinator: coordinator,
		ctx:         ctx,
	}
}

func (s *SignService) Work() {
	logger.DefaultLogStartWork("SignService")
	defer logger.DefaultLogFinishWork("SignService", nil)

	for {
		s.ExecuteSign()
		time.Sleep(3 * time.Second)
	}
}

func (s *SignService) ExecuteSign() {
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
	s.execute(dkg)
}

func (s *SignService) execute(dkg *coordinator.DKG) {
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

	s.logMessage(fmt.Sprintf("Processing pegout ID: %x", unsignedPegout.ID))
	s.logMessage(fmt.Sprintf("Pegout address: %s", unsignedPegout.PegoutAddress))

	valKey, err := s.validator.GetValidatorKey(dkg)
	if err != nil {
		s.logError(fmt.Errorf("failed to get validator key: %w", err))
		return
	}

	if valKey == nil {
		s.logError(fmt.Errorf("Oracle is not a validator. Cannot participate in signing pegout %x", unsignedPegout.ID))
		return
	}

	if err := s.coordinator.ConnectSigner(s.validator.GetSigner(valKey.ValidatorID)); err != nil {
		s.logError(fmt.Errorf("failed to connect to coordinator: %w", err))
		return
	}

	minSigners := int(math.Floor(float64(dkg.MaxSigners) * 2 / 3))

	// Execute signing steps
	if s.doCommit(valKey, pegoutTxID, &pegoutTx, minSigners) & 
		s.doSign(ctx, valKey, pegoutTxID, &pegoutTx, minSigners) & {
		
	}

	if err :=  err != nil {
		return fmt.Errorf("sign failed: %w", err)
	}

	if err := s.doAggregate(ctx, valKey, pegoutTxID, &pegoutTx); err != nil {
		return fmt.Errorf("aggregate failed: %w", err)
	}

	return nil
}

func (s *SignService) doCommit(
	validatorKey *validator.ValidatorKey,
	pegoutID uint64,
	pegoutRecord *coordinator.PegoutRecord,
	minSigners int,
) bool {
	s.logCommit(pegoutID)
	identifier := validatorKey.ValidatorKey

	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if pegoutRecord.HasCommitment(identifier) {
		if pegoutRecord.CommitmentsCount() >= minSigners {
			s.logMessage(fmt.Sprintf("Moving to signing phase for pegout %x", pegoutID))
			return true
		}
		return false
	}

	nonce := s.keyStore.LoadNonce(pegoutAddressStr)
	commitments := s.keyStore.LoadCommitments(pegoutAddressStr)

	if nonce == nil || commitments == nil {
		if nonce == nil && commitments == nil {
			nonce, commitments, err := s.dkgClient.Commit(pegoutRecord.InternalKey)
			if err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			// TODO
			//s.keyStore.StoreNonce(pegoutAddressStr, nonce)
			//s.keyStore.StoreCommitments(pegoutAddressStr, commitments)
		} else {
			s.logMessage(fmt.Errorf("problem with saved nonce or commitments for %s", pegoutAddressStr))
			return false
		}
	}

	if err := s.coordinator.SendCommitments(ctx, &CommitmentRequest{
		PegoutID:     pegoutID,
		ValidatorIdx: validatorKey.ValidatorIdx,
		Identifier:   identifier,
		Commitments:  commitments,
	}); err != nil {
		s.logMessage(fmt.Errorf("failed to send commitments: %w", err))
		return false
	}

	s.logMessage(fmt.Sprintf("Commit sent for pegout %x", pegoutID))
	return true
}

func (s *SignService) doSign(
	ctx context.Context, 
	validatorKey *ValidatorKey, 
	pegoutID uint64, 
	pegoutRecord *PegoutRecord, 
	minSigners int,
) bool {
	s.logMessage(fmt.Sprintf("Sign pegout %x", pegoutID))
	identifier := validatorKey.ValidatorKey

	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if !pegoutRecord.HasCommitment(identifier) {
		s.logMessage(fmt.Sprintf("Oracle didn't send commitment and cannot participate in signing for pegout %x", pegoutID))
		return false
	}

	commitsArr := make([]CommitmentPackage, 0)
	for key, commitment := range pegoutRecord.Commitments {
		commitsArr = append(commitsArr, CommitmentPackage{
			Identifier: key,
			Package:    commitment,
		})
	}

	pegoutTxContract := s.tonService.OpenPegoutTx(pegoutRecord.PegoutAddress)
	signHashes, err := pegoutTxContract.GetSigningHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get signing hashes: %w", err)
	}

	if len(signHashes) == 0 {
		s.logMessage(fmt.Sprintf(("NO signing hashes in pegout %x", pegoutID)
		return nil
	}

	signPkgs := s.keyStore.LoadTempArray(fmt.Sprintf("pkgs_%s", pegoutAddressStr))
	if signPkgs == nil {
		signPkgs = make([][]byte, 0, len(signHashes))
		for _, hash := range signHashes {
			signPkg, err := CreateSigningPackage(commitsArr, hash)
			if err != nil {
				return fmt.Errorf("failed to create signing package: %w", err)
			}
			signPkgs = append(signPkgs, signPkg)
		}
		s.keyStore.StoreTempArray(fmt.Sprintf("pkgs_%s", pegoutAddressStr), signPkgs)
	}

	if !pegoutRecord.HasSigningShare(identifier) {
		signShares := s.keyStore.LoadTempArray(fmt.Sprintf("shares_%s", pegoutAddressStr))
		if signShares == nil {
			nonce := s.keyStore.LoadTemp(fmt.Sprintf("nonce_%s", pegoutAddressStr))
			if nonce == nil {
				return fmt.Errorf("signing nonce is undefined")
			}

			txParts, err := pegoutTxContract.GetTxParts(ctx)
			if err != nil {
				return fmt.Errorf("failed to get tx parts: %w", err)
			}

			signShares = make([][]byte, 0, len(signHashes))
			for i := range signHashes {
				signShare, err := s.dkgService.Sign(ctx, pegoutRecord.InternalKey, signPkgs[i], nonce, txParts.Inputs[i].TaprootMerkleRoot)
				if err != nil {
					return fmt.Errorf("failed to sign: %w", err)
				}
				signShares = append(signShares, signShare)
			}
			s.keyStore.StoreTempArray(fmt.Sprintf("shares_%s", pegoutAddressStr), signShares)
		}

		if err := s.coordinator.SendSigningShare(ctx, &SigningShareRequest{
			PegoutID:      pegoutID,
			ValidatorIdx:  validatorKey.ValidatorIdx,
			Identifier:    identifier,
			SigningShares: signShares,
			Lifetime:      30,
		}); err != nil {
			return fmt.Errorf("failed to send signing share: %w", err)
		}

		s.logMessage(fmt.Sprintf(("Signing share sent for pegout %x", pegoutID)
		return nil
	}

	if pegoutRecord.SigningSharesCount() >= minSigners {
		s.logMessage(fmt.Sprintf(("Moving to aggregation phase for pegout %x", pegoutID)
		return nil
	}

	return nil
}

func (s *SignService) doAggregate(ctx context.Context, validatorKey *ValidatorKey, pegoutID uint64, pegoutRecord *PegoutRecord) error {
	s.logMessage(fmt.Sprintf(("Aggregate sign shares for pegout %x", pegoutID)
	identifier := validatorKey.ValidatorKey

	pegoutTxContract := s.tonService.OpenPegoutTx(pegoutRecord.PegoutAddress)

	txParts, err := pegoutTxContract.GetTxParts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tx parts: %w", err)
	}

	if len(txParts.Signatures) > 0 {
		s.logger.Println("Completed. Signature already exists")
		return nil
	}

	prevDkg, err := s.coordinator.GetPrevDKG(ctx)
	if err != nil {
		return fmt.Errorf("failed to get previous DKG: %w", err)
	}

	pubkeyPackage := prevDkg.R3Package.PubkeyData.PubkeyPackage

	sharesArr := make([]SigningShare, 0)
	for key, shares := range pegoutRecord.SigningShares {
		for idx, share := range shares {
			sharesArr = append(sharesArr, SigningShare{
				Identifier: key,
				Package:    share,
				Index:      idx,
			})
		}
	}

	signPkgs := s.keyStore.LoadTempArray(fmt.Sprintf("pkgs_%s", pegoutRecord.PegoutAddress.String()))
	if signPkgs == nil {
		s.logger.Println("Signing packages array is empty")
		return nil
	}

	signatures := make([][]byte, 0, len(signPkgs))
	for i, pkg := range signPkgs {
		filteredShares := filterSharesByIndex(sharesArr, i)
		signature, err := Aggregate(pkg, filteredShares, pubkeyPackage, txParts.Inputs[i].TaprootMerkleRoot)
		if err != nil {
			return fmt.Errorf("failed to aggregate signatures: %w", err)
		}
		signatures = append(signatures, signature)
	}

	if err := s.coordinator.SendSignatures(ctx, &SignaturesRequest{
		PegoutID:     pegoutID,
		ValidatorIdx: validatorKey.ValidatorIdx,
		Identifier:   identifier,
		Signatures:   signatures,
		Lifetime:     30,
	}); err != nil {
		return fmt.Errorf("failed to send signatures: %w", err)
	}

	s.logMessage(fmt.Sprintf(("Signature sent for pegout %x", pegoutID)
	return nil
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
