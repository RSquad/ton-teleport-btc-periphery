package signing

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
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
		SendCommitmentsFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
			panic("mock out the SendCommitments method")
		},
		SendDKGClaimFunc: func(validatorIdx uint16, dkgUntil int64, culpritIdx uint16) (*tlb.Transaction, error) {
			panic("mock out the SendDKGClaim method")
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
		SendSignaturesFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
			if validatorIdx == 1 {
				// Emulate an error (err::different_pegout_signatures = 168) during the coordinator contract call
				return nil, errors.New("...exitcode=168...")
			}

			// Emulate a successful send signatures to the coordinator contract
			return nil, nil
		},
		SendSigningClaimFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, culpritIdx uint16) (*tlb.Transaction, error) {
			if validatorIdx != 1 {
				panic(fmt.Sprintf("Wrong validatorIdx: expected 1, but got %d", validatorIdx))
			}

			if culpritIdx != 0 {
				panic(fmt.Sprintf("Wrong culpritIdx: expected 0, but got %d", culpritIdx))
			}

			// Emulate a successful send claim to the coordinator contract
			return nil, nil
		},
		SendSigningShareFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signingShares [][]byte) (*tlb.Transaction, error) {
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

func TestPegoutUntil(t *testing.T) {
	pegout := &CachedPegout{
		ID:            1,
		addrStr:       "",
		inputs:        []pegoutcontract.TxInput{},
		tx:            nil,
		signingHashes: [][]byte{},
		artifacts: &coordinator.PegoutRecord{
			ExpiredAt:  time.Unix(100000000, 0),
			ClaimsMask: big.NewInt(0),
		},
	}

	validateExpiredAt := func(pegout *CachedPegout, pegoutUntil int64) {
		if pegout.artifacts.ExpiredAt.Unix() != pegoutUntil {
			t.Fatalf("Wrong pegoutUntil: expected %d, but got %d", pegout.artifacts.ExpiredAt.Unix(), pegoutUntil)
		}
	}

	coordinator := &coordinator.CoordinatorMock{
		SendCommitmentsFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
			validateExpiredAt(pegout, pegoutUntil)
			return nil, nil
		},
		SendSigningShareFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signingShares [][]byte) (*tlb.Transaction, error) {
			validateExpiredAt(pegout, pegoutUntil)
			return nil, nil
		},
		SendSignaturesFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
			validateExpiredAt(pegout, pegoutUntil)
			return nil, nil
		},
		SendSigningClaimFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, culpritIdx uint16) (*tlb.Transaction, error) {
			validateExpiredAt(pegout, pegoutUntil)
			return nil, nil
		},
	}
	service := NewService(nil, coordinator, nil, 0)
	validatorIdx := uint16(0)

	commitments := make([]byte, 32)
	rand.Read(commitments)

	service.sendCommitments(pegout, validatorIdx)

	signShares := make([][]byte, 1)
	signShares[0] = make([]byte, 32)
	rand.Read(signShares[0])

	service.sendSigningShares(pegout, validatorIdx, signShares)

	signatures := make([][]byte, 1)
	signatures[0] = make([]byte, 64)
	rand.Read(signatures[0])

	service.SendSignatures(pegout, validatorIdx, signatures)

	service.executeClaim(pegout, validatorIdx, 1)
}

func generateFrostKey(minSigners uint16, maxSigners uint16) (map[uint16][]byte, []byte) {
	r1Secrets := make(map[frost.Identifier]uintptr)
	r2Secrets := make(map[frost.Identifier]uintptr)
	receivedR1Packages := make(map[frost.Identifier]map[frost.Identifier]frost.Package)
	receivedR2Packages := make(map[frost.Identifier]map[frost.Identifier]frost.Package)

	frostIdentifiers := make([]frost.Identifier, maxSigners)
	for i := uint16(0); i < maxSigners; i++ {
		frostIdentifiers[i] = helpers.ValidatorIdxToFrost(i)
	}

	for _, id := range frostIdentifiers {
		pkg, secret, err := frost.DkgPart1(id, minSigners, maxSigners)
		if err != nil {
			panic(err)
		}
		r1Secrets[id] = secret
		for _, destId := range frostIdentifiers {
			if id != destId {
				pkgs, ok := receivedR1Packages[destId]
				if !ok {
					pkgs = make(map[frost.Identifier]frost.Package)
				}
				pkgs[id] = frost.NewPackage(pkg)
				receivedR1Packages[destId] = pkgs
			}
		}
	}

	for _, id := range frostIdentifiers {
		secret := r1Secrets[id]
		r2Packages, r2secret, _, err := frost.DkgPart2(secret, receivedR1Packages[id])
		if err != nil {
			panic(err)
		}
		frost.FreeR1Secret(secret)
		delete(r1Secrets, id)
		r2Secrets[id] = r2secret
		for destId, r2package := range r2Packages {
			pkgs, ok := receivedR2Packages[destId]
			if !ok {
				pkgs = make(map[frost.Identifier]frost.Package)
			}
			pkgs[id] = r2package
			receivedR2Packages[destId] = pkgs
		}
	}

	keyPackages := make(map[uint16][]byte)
	publicKeyPackage := []byte{}
	for _, id := range frostIdentifiers {
		keyPackage, pubPackage, _, err := frost.DkgPart3(
			r2Secrets[id],
			receivedR1Packages[id],
			receivedR2Packages[id],
		)
		frost.FreeR2Secret(r2Secrets[id])
		delete(r2Secrets, id)
		if err != nil {
			panic(err)
		}
		keyPackages[helpers.FrostToValidatorIdx(id)] = keyPackage
		publicKeyPackage = pubPackage
	}

	return keyPackages, publicKeyPackage
}

func TestGenerateCommitmentsForEachInput(t *testing.T) {
	maxSigners := uint16(3)
	minSigners := uint16(2)

	// Generate frost group key
	keyPackages, publicKeyPackage := generateFrostKey(minSigners, maxSigners)
	groupPublicKey, err := frost.ExtractPublicKeyFromPackage(publicKeyPackage)
	if err != nil {
		t.Fatal(err)
	}

	// Prepare cached pegout
	pegout := &CachedPegout{
		ID:      1,
		addrStr: address.NewAddress(0, 0, make([]byte, 32)).String(),
		inputs: []pegoutcontract.TxInput{{
			TxHash: make([]byte, 32),
			Data: &pegoutcontract.TxPartsInput{
				Amount: big.NewInt(10000),
				Index:  0,
			},
		}, {
			TxHash: make([]byte, 32),
			Data: &pegoutcontract.TxPartsInput{
				Amount: big.NewInt(20000),
				Index:  1,
			},
		}},
		tx: &pegoutcontract.TxParts{
			InternalKey: groupPublicKey,
		},
		artifacts: &coordinator.PegoutRecord{
			Commitments:     map[uint16][]byte{},
			CommitmentsMask: make([]byte, 32),
			MaxSigners:      maxSigners,
			ExpiredAt:       time.Now().Add(time.Minute),
			SigningMask:     big.NewInt(7),
			Signatures: coordinator.PegoutSignatures{
				Mask:  big.NewInt(0),
				Count: 0,
				Hash:  make([]byte, 32),
			},
		},
		signingHashes: [][]byte{
			make([]byte, 32),
			make([]byte, 32),
		},
	}
	// Create keystore mock
	keystore := &keystore.KeystoreMock{
		LoadSecretFunc: func(pubKey []byte) []byte {
			return keyPackages[0]
		},
	}

	coordinator := &coordinator.CoordinatorMock{
		SendCommitmentsFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
			if validatorIdx != 0 {
				t.Fatal("validatorIdx should be 0")
			}
			pegout.artifacts.Commitments[validatorIdx] = commitments
			pegout.artifacts.CommitmentsMask = []byte{1}
			pegout.artifacts.SigningMask = big.NewInt(1)
			return nil, nil
		},
		GetPrevDKGFunc: func() (*coordinator.DKG, error) {
			return &coordinator.DKG{
				State: coordinator.DKGStateFinished,
				VSet: coordinator.VSet{
					0: make([]byte, 32),
					1: make([]byte, 32),
					2: make([]byte, 32),
				},
				MaxSigners: maxSigners,
				VSetMask:   big.NewInt(1<<maxSigners - 1),
				SessionKeys: &coordinator.SessionKeys{
					PubKeys: coordinator.SessionPubKeys{
						0: make([]byte, 32),
						1: make([]byte, 32),
						2: make([]byte, 32),
					},
				},
				Until: time.Now().Add(time.Hour),
				R3: &coordinator.DKGR3{
					Mask:  big.NewInt(1<<maxSigners - 1),
					Count: maxSigners,
					Data: &coordinator.PubkeyData{
						PubkeyPackage: publicKeyPackage,
						InternalKey:   groupPublicKey,
					},
				},
			}, nil
		},
		SendSignaturesFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
			return nil, nil
		},
	}

	// Create service instance
	service := NewService(keystore, coordinator, nil, 0)
	service.cachedPegout = pegout
	validatorIdx := uint16(0)

	t.Run("doCommit", func(t *testing.T) {
		service.doCommit(validatorIdx, minSigners)
		t.Run("Check that commitment is set", func(t *testing.T) {
			if !pegout.artifacts.HasCommitment(validatorIdx) {
				t.Fatal("commitment should be set")
			}
		})
		var commitments [][]byte
		t.Run("Commitments are deserialized correctly", func(t *testing.T) {
			commitments, err = helpers.DeserializeCommitments(pegout.artifacts.Commitments[validatorIdx], len(pegout.inputs), helpers.FrostCommitmentLength)
			if err != nil {
				t.Fatal(err)
			}
			if len(commitments) != len(pegout.inputs) {
				t.Fatal("commitments count should be equal to the number of inputs")
			}
		})
		t.Run("Commitments are different", func(t *testing.T) {
			for i, localCommitment := range pegout.commitments {
				for j, commit := range pegout.commitments {
					if i != j && bytes.Equal(localCommitment, commit) {
						t.Fatal("commitments should be different")
					}
				}
			}
		})
		t.Run("Nonces are different", func(t *testing.T) {
			for i, localNonce := range pegout.nonces {
				for j, nonce := range pegout.nonces {
					if i != j && bytes.Equal(localNonce, nonce) {
						t.Fatal("nonces should be different")
					}
				}
			}
		})

		t.Run("Local and coordinator commitments are equal", func(t *testing.T) {
			for i, localCommitment := range pegout.commitments {
				if !bytes.Equal(localCommitment, commitments[i]) {
					t.Fatal("commitments should be equal")
				}
			}
		})
	})

	generateCommitments := func(idx uint16, inputs uint16) [][]byte {
		commitments := make([][]byte, inputs)
		var err error
		for i := range commitments {
			_, commitments[i], err = frost.Commit(frost.NewPackage(keyPackages[idx]))
			if err != nil {
				t.Fatal(err)
			}
		}
		return commitments
	}

	t.Run("signInput: deserialize commitments", func(t *testing.T) {
		commitmentsFor1Serialized, err := helpers.SerializeCommitments(generateCommitments(1, 2), helpers.FrostCommitmentLength)
		if err != nil {
			t.Fatal(err)
		}
		commitmentsFor2Serialized, err := helpers.SerializeCommitments(generateCommitments(2, 2), helpers.FrostCommitmentLength)
		if err != nil {
			t.Fatal(err)
		}
		pegout.artifacts.Commitments[1] = commitmentsFor1Serialized
		pegout.artifacts.Commitments[2] = commitmentsFor2Serialized

		t.Run("SignInput 0", func(t *testing.T) {
			_, err = service.SignInput(validatorIdx, 0)
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("SignInput 1", func(t *testing.T) {
			_, err = service.SignInput(validatorIdx, 1)
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("Cleanup nonces in doAggregate", func(t *testing.T) {
		if service.cachedPegout.nonces == nil {
			t.Fatal("inputs should be set")
		}
		service.cachedPegout.inputs = make([]pegoutcontract.TxInput, 0)
		service.doAggregate(validatorIdx, publicKeyPackage)

		if service.cachedPegout.nonces != nil {
			t.Fatal("nonces should be nil")
		}
		// call again to be sure the code doesn't panic
		service.cleanupNonces()
	})
}
