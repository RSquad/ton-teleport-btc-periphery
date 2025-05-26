package signing

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	tonutils "github.com/xssnick/tonutils-go/ton"
)

func newCoordinatorContractMock() *coordinator.CoordinatorMock {
	return &coordinator.CoordinatorMock{
		ConnectSignerFunc: func(signerMoqParam signer.Signer) {
			panic("mock out the ConnectSigner method")
		},
		GetAddrFunc: func() *address.Address {
			panic("mock out the GetAddr method")
		},
		GetDkgFunc: func(block *tonutils.BlockIDExt) (*coordinator.DKG, error) {
			panic("mock out the GetDkg method")
		},
		GetPrevDKGFunc: func() (*coordinator.DKG, error) {
			panic("mock out the GetPrevDKG method")
		},
		GetUnsignedPegoutsFunc: func() ([]coordinator.PegoutRecord, error) {
			panic("mock out the GetUnsignedPegouts method")
		},
		SendCommitmentsFunc: func(pegoutID uint64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
			panic("mock out the SendCommitments method")
		},
		SendDKGClaimFunc: func(validatorIdx uint16, dkgUntil int64, culpritIdx uint16) (*tlb.Transaction, error) {
			panic("mock out the SendDKGClaim method")
		},
		SendPubkeyPackageFunc: func(validatorIdx uint16, dkgUntil int64, sessionPublicKey []byte, pubkeyPackage []byte) (*tlb.Transaction, error) {
			panic("mock out the SendPubkeyPackage method")
		},
		SendResetPegoutSigningFunc: func(pegoutID uint64, validatorIdx uint16) (*tlb.Transaction, error) {
			panic("mock out the SendResetPegoutSigning method")
		},
		SendRound1Func: func(validatorIdx uint16, dkgUntil int64, round1Package []byte, r2PublicX25519 *[32]byte) (*tlb.Transaction, error) {
			panic("mock out the SendRound1 method")
		},
		SendRound2Func: func(validatorIdx uint16, dkgUntil int64, round2Packages []byte) (*tlb.Transaction, error) {
			panic("mock out the SendRound2 method")
		},
		SendSignaturesFunc: func(pegoutID uint64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
			if validatorIdx == 1 {
				// Emulate an error (err::different_pegout_signatures = 168) during the coordinator contract call
				return nil, errors.New("...exitcode=168...")
			}

			// Emulate a successful send signatures to the coordinator contract
			return nil, nil
		},
		SendSigningClaimFunc: func(pegoutID uint64, validatorIdx uint16, culpritIdx uint16) (*tlb.Transaction, error) {
			if validatorIdx != 1 {
				panic(fmt.Sprintf("Wrong validatorIdx: expected 1, but got %d", validatorIdx))
			}

			if culpritIdx != 0 {
				panic(fmt.Sprintf("Wrong culpritIdx: expected 0, but got %d", culpritIdx))
			}

			// Emulate a successful send claim to the coordinator contract
			return nil, nil
		},
		SendSigningShareFunc: func(pegoutID uint64, validatorIdx uint16, signingShares [][]byte) (*tlb.Transaction, error) {
			panic("mock out the SendSigningShare method")
		},
		SendStartDKGFunc: func() (*tlb.Transaction, error) {
			panic("mock out the SendStartDKG method")
		},
	}
}

func newCachedPegout() *CachedPegout {
	return &CachedPegout{
		ID:            0,
		addrStr:       "",
		inputs:        []pegoutcontract.TxInput{},
		tx:            nil,
		signingHashes: [][]byte{},
		artifacts: &coordinator.PegoutRecord{
			Signatures: coordinator.PegoutSignatures{
				Mask:  big.NewInt(0b1),
				Count: 2,
				Hash:  []byte{},
			},
			ClaimsMask: big.NewInt(0),
		},
	}
}

func TestSignService_SendSignatures_Success(t *testing.T) {
	coordinatorContractMock := newCoordinatorContractMock()
	service := NewService(nil, coordinatorContractMock, nil, 0)

	validatorIdx := uint16(0)
	pegout := newCachedPegout()

	resCode := service.SendSignatures(pegout, validatorIdx, nil)
	if resCode != 0 {
		t.Fatalf("expected resCode = 0, got %d", resCode)
	}
}

func TestSignService_SendSignatures_ERROR_DifferentPegoutSignatures(t *testing.T) {
	coordinatorContractMock := newCoordinatorContractMock()
	service := NewService(nil, coordinatorContractMock, nil, 0)

	validatorIdx := uint16(1)
	pegout := newCachedPegout()

	resCode := service.SendSignatures(pegout, validatorIdx, nil)
	if resCode != helpers.DifferentPegoutSignatures {
		t.Fatalf("expected resCode = %d (DifferentPegoutSignatures), got %d", helpers.DifferentPegoutSignatures, resCode)
	}
}
