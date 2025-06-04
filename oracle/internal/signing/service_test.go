package signing

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

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
		StoreNonceFunc: func(name string, nonce []byte) error {
			t.Fatal("StoreNonceFunc should not be called")
			return nil
		},
		LoadNonceFunc: func(name string) []byte {
			t.Fatal("LoadNonceFunc should not be called")
			return nil
		},
	}

	coordinator := &coordinator.CoordinatorMock{
		SendCommitmentsFunc: func(pegoutID uint64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
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
		SendSignaturesFunc: func(pegoutID uint64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
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

	t.Run("doSign: deserialize commitments", func(t *testing.T) {
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
