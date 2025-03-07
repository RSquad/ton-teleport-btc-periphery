package signer

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/dkg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
)

type SignService struct {
	logger           *log.Logger
	mu               sync.Mutex
	dkgService       *dkg.DkgService
	config           *cfg.Cfg
	keyStore         keystore.Keystore
	coordinator      *coordinatorcontract.CoordinatorContract
	validatorService ValidatorService
	ctx              context.Context
}

func NewSignService(
	config *cfg.Cfg,
	dkgService *dkg.DkgService,
	keyStore keystore.Keystore,
	validatorService ValidatorService,
	coordinator *coordinatorcontract.CoordinatorContract,
	ctx context.Context,
) (*SignService, error) {
	return &SignService{
		logger:           log.New(log.Writer(), "SignService: ", log.LstdFlags),
		config:           config,
		dkgService:       dkgService,
		keyStore:         keyStore,
		validatorService: validatorService,
		coordinator:      coordinator,
		ctx:              ctx,
	}, nil
}

func (s *SignService) ExecuteSign() {
	defer func() {
		s.logger.Println("Cron Job completed")
	}()

	s.logger.Println("Cron Job started")

	dkg, err := s.coordinator.GetPrevDKG()
	if err != nil {
		s.logger.Printf("failed to get previous DKG: %w", err)
		return
	}

	if dkg == nil {
		s.logger.Println("DKG not yet completed")
		return
	}
	s.execute(dkg)
}

func (s *SignService) execute(dkg *coordinatorcontract.DKG) error {
	pegoutRecords, err := s.coordinator.GetUnsignedPegouts()
	if err != nil {
		return fmt.Errorf("failed to get unsigned pegouts: %w", err)
	}

	if len(pegoutRecords) == 0 {
		s.logger.Println("No sign requests")
		return nil
	}

	s.logger.Printf("%d signing requests", len(pegoutRecords))

	// Get first pegout record
	var pegoutTxID uint64
	var pegoutTx PegoutRecord
	for id, tx := range pegoutRecords {
		pegoutTxID = id
		pegoutTx = tx
		break
	}

	s.logger.Printf("Processing pegout ID: %x", pegoutTxID)
	s.logger.Printf("Pegout address: %s", pegoutTx.PegoutAddress)

	valKey, err := s.validatorService.GetValidatorKey(ctx, dkg)
	if err != nil {
		return fmt.Errorf("failed to get validator key: %w", err)
	}

	if valKey == nil {
		s.logger.Printf("Oracle is not a validator. Cannot participate in signing pegout %x", pegoutTxID)
		return nil
	}

	if err := s.coordinator.Connect(ctx, s.validatorService.GetSigner(valKey.ValidatorID)); err != nil {
		return fmt.Errorf("failed to connect to coordinator: %w", err)
	}

	minSigners := int(math.Floor(float64(dkg.MaxSigners) * 2 / 3))

	// Execute signing steps
	if err := s.doCommit(ctx, valKey, pegoutTxID, &pegoutTx, minSigners); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	if err := s.doSign(ctx, valKey, pegoutTxID, &pegoutTx, minSigners); err != nil {
		return fmt.Errorf("sign failed: %w", err)
	}

	if err := s.doAggregate(ctx, valKey, pegoutTxID, &pegoutTx); err != nil {
		return fmt.Errorf("aggregate failed: %w", err)
	}

	return nil
}

func (s *SignService) doCommit(ctx context.Context, validatorKey *ValidatorKey, pegoutID uint64, pegoutRecord *PegoutRecord, minSigners int) error {
	s.logger.Printf("Commit pegout %x", pegoutID)
	identifier := validatorKey.ValidatorKey

	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if pegoutRecord.HasCommitment(identifier) {
		if pegoutRecord.CommitmentsCount() >= minSigners {
			s.logger.Printf("Moving to signing phase for pegout %x", pegoutID)
			return nil
		}
		return nil
	}

	nonce := s.keyStore.LoadTemp(fmt.Sprintf("nonce_%s", pegoutAddressStr))
	commitments := s.keyStore.LoadTemp(fmt.Sprintf("commitments_%s", pegoutAddressStr))

	if nonce == nil || commitments == nil {
		if nonce == nil && commitments == nil {
			commitResult, err := s.dkgService.Commit(ctx, pegoutRecord.InternalKey)
			if err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			nonce = commitResult.Nonce
			commitments = commitResult.Commitments

			s.keyStore.StoreTemp(fmt.Sprintf("nonce_%s", pegoutAddressStr), nonce)
			s.keyStore.StoreTemp(fmt.Sprintf("commitments_%s", pegoutAddressStr), commitments)
		} else {
			return fmt.Errorf("problem with saved nonce or commitments for %s", pegoutAddressStr)
		}
	}

	if err := s.coordinator.SendCommitments(ctx, &CommitmentRequest{
		PegoutID:     pegoutID,
		ValidatorIdx: validatorKey.ValidatorIdx,
		Identifier:   identifier,
		Commitments:  commitments,
		Lifetime:     30,
	}); err != nil {
		return fmt.Errorf("failed to send commitments: %w", err)
	}

	s.logger.Printf("Commit sent for pegout %x", pegoutID)
	return nil
}

func (s *SignService) doSign(ctx context.Context, validatorKey *ValidatorKey, pegoutID uint64, pegoutRecord *PegoutRecord, minSigners int) error {
	s.logger.Printf("Sign pegout %x", pegoutID)
	identifier := validatorKey.ValidatorKey

	pegoutAddressStr := pegoutRecord.PegoutAddress.String()

	if !pegoutRecord.HasCommitment(identifier) {
		s.logger.Printf("Oracle didn't send commitment and cannot participate in signing for pegout %x", pegoutID)
		return nil
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
		s.logger.Printf("NO signing hashes in pegout %x", pegoutID)
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

		s.logger.Printf("Signing share sent for pegout %x", pegoutID)
		return nil
	}

	if pegoutRecord.SigningSharesCount() >= minSigners {
		s.logger.Printf("Moving to aggregation phase for pegout %x", pegoutID)
		return nil
	}

	return nil
}

func (s *SignService) doAggregate(ctx context.Context, validatorKey *ValidatorKey, pegoutID uint64, pegoutRecord *PegoutRecord) error {
	s.logger.Printf("Aggregate sign shares for pegout %x", pegoutID)
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

	s.logger.Printf("Signature sent for pegout %x", pegoutID)
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
