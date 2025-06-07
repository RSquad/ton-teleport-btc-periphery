package signing

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
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

	service.sendCommitments(pegout, validatorIdx, commitments)

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

func deepCopy2dSlice[T any](slice [][]T) [][]T {
	copied := make([][]T, len(slice))
	for i, row := range slice {
		copied[i] = make([]T, len(row))
		copy(copied[i], row)
	}
	return copied
}

func TestNonceAndCommitmentCleanupOnExpiredAtChange(t *testing.T) {
	maxSigners := uint16(3)
	minSigners := uint16(2)

	// Generate frost group key
	keyPackages, publicKeyPackage := generateFrostKey(minSigners, maxSigners)
	groupPublicKey, err := frost.ExtractPublicKeyFromPackage(publicKeyPackage)
	if err != nil {
		t.Fatal(err)
	}

	originalExpiredAt := time.Now().Add(time.Hour)

	unsignedPegout := &coordinator.PegoutRecord{
		ID:              1,
		Commitments:     map[uint16][]byte{},
		CommitmentsMask: make([]byte, 32),
		MaxSigners:      maxSigners,
		ExpiredAt:       originalExpiredAt,
		SigningMask:     big.NewInt(7),
	}
	// Prepare cached pegout with initial ExpiredAt
	pegout := &CachedPegout{
		ID:      1,
		addrStr: address.NewAddress(0, 0, make([]byte, 32)).String(),
		inputs: []pegoutcontract.TxInput{{
			TxHash: make([]byte, 32),
			Data: &pegoutcontract.TxPartsInput{
				Amount: big.NewInt(10000),
				Index:  0,
			},
		}},
		tx: &pegoutcontract.TxParts{
			InternalKey: groupPublicKey,
		},
		artifacts: &coordinator.PegoutRecord{
			ID:              1,
			Commitments:     map[uint16][]byte{},
			CommitmentsMask: make([]byte, 32),
			MaxSigners:      maxSigners,
			ExpiredAt:       originalExpiredAt,
			SigningMask:     big.NewInt(7),
		},
		commitments: nil,
		nonces:      nil,
		signingHashes: [][]byte{
			make([]byte, 32),
		},
	}

	dkgUntil := time.Now().Add(time.Hour)
	publicKey, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	// Create keystore mock
	cleanupCalled := false
	keystore := &keystore.KeystoreMock{
		LoadSecretFunc: func(pubKey []byte) []byte {
			return keyPackages[0]
		},
		CleanupFunc: func() {
			cleanupCalled = true
		},
		LoadSessionFunc: func(dkgUntil int64) []byte {
			return secret
		},
	}

	prevDKG := &coordinator.DKG{
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
				0: publicKey[:],
				1: make([]byte, 32),
				2: make([]byte, 32),
			},
		},
		Until: dkgUntil,
		R3: &coordinator.DKGR3{
			Mask:  big.NewInt(1<<maxSigners - 1),
			Count: maxSigners,
			Data: &coordinator.PubkeyData{
				PubkeyPackage: publicKeyPackage,
				InternalKey:   groupPublicKey,
			},
		},
	}

	coordinator := &coordinator.CoordinatorMock{
		SendCommitmentsFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, commitments []byte) (*tlb.Transaction, error) {
			pegout.artifacts.Commitments[validatorIdx] = commitments
			mask := big.NewInt(0).SetBytes(pegout.artifacts.CommitmentsMask)
			mask.SetBit(mask, int(validatorIdx), 1)
			pegout.artifacts.CommitmentsMask = mask.FillBytes(make([]byte, 32))
			return nil, nil
		},
		GetUnsignedPegoutsFunc: func() ([]coordinator.PegoutRecord, error) {
			return []coordinator.PegoutRecord{*unsignedPegout}, nil
		},
		GetPrevDKGFunc: func() (*coordinator.DKG, error) {
			return prevDKG, nil
		},
		SendSignaturesFunc: func(pegoutID uint64, pegoutUntil int64, validatorIdx uint16, signatures [][]byte) (*tlb.Transaction, error) {
			return nil, nil
		},
		ConnectSignerFunc: func(signer signer.Signer) {
		},
	}

	// Create service instance
	service := NewService(keystore, coordinator, nil, 0)
	service.cachedPegout = pegout
	validatorIdx := uint16(0)

	t.Run("Generate initial commitments and nonces", func(t *testing.T) {
		// Generate commitments to populate nonces and commitments
		service.execute(context.Background(), prevDKG)

		// Verify commitments and nonces are generated
		if len(pegout.commitments) == 0 {
			t.Fatal("commitments should be generated")
		}
		if len(pegout.nonces) == 0 {
			t.Fatal("nonces should be generated")
		}
		if !pegout.artifacts.HasCommitment(validatorIdx) {
			t.Fatal("commitment should be set in artifacts")
		}
	})

	t.Run("Run execute again to be sure the nonces and commitments are not regenerated", func(t *testing.T) {
		originalCommitments := deepCopy2dSlice(pegout.commitments)
		originalNonces := deepCopy2dSlice(pegout.nonces)

		service.execute(context.Background(), prevDKG)
		for i, newNonce := range pegout.nonces {
			if !bytes.Equal(newNonce, originalNonces[i]) {
				t.Fatalf("nonces should be the same after execute. Original: %x, New: %x", originalNonces[i], newNonce)
			}
		}
		for i, newCommitment := range pegout.commitments {
			if !bytes.Equal(newCommitment, originalCommitments[i]) {
				t.Fatalf("commitments should be the same after execute. Original: %x, New: %x", originalCommitments[i], newCommitment)
			}
		}
	})

	t.Run("ExpiredAt change should drop nonces and commitments", func(t *testing.T) {
		// Store original values for comparison
		originalCommitments := deepCopy2dSlice(pegout.commitments)
		originalNonces := deepCopy2dSlice(pegout.nonces)

		// Change ExpiredAt to simulate restart/time change scenario
		unsignedPegout.ExpiredAt = originalExpiredAt.Add(time.Hour)
		// reset commitments and commitments mask
		unsignedPegout.Commitments = map[uint16][]byte{}
		unsignedPegout.CommitmentsMask = make([]byte, 32)

		// Verify nonces and commitments exist before cleanup
		if pegout.commitments == nil {
			t.Fatal("commitments should exist before cleanup")
		}
		if pegout.nonces == nil {
			t.Fatal("nonces should exist before cleanup")
		}

		service.execute(context.Background(), prevDKG)

		// Verify that nonces and commitments were regenerated (different from original)
		if len(pegout.commitments) == 0 {
			t.Fatal("new commitments should be generated after ExpiredAt change")
		}
		if len(pegout.nonces) == 0 {
			t.Fatal("new nonces should be generated after ExpiredAt change")
		}

		if !cleanupCalled {
			t.Fatal("cleanup should be called")
		}

		// Verify that new nonces are different from original ones (preventing reuse)
		for i, newNonce := range pegout.nonces {
			if i < len(originalNonces) && bytes.Equal(newNonce, originalNonces[i]) {
				t.Fatalf("nonces should be different after ExpiredAt change to prevent reuse attacks. Original: %x, New: %x", originalNonces[i], newNonce)
			}
		}

		// Verify that new commitments are different from original ones
		for i, newCommitment := range pegout.commitments {
			if i < len(originalCommitments) && bytes.Equal(newCommitment, originalCommitments[i]) {
				t.Fatal("commitments should be different after ExpiredAt change")
			}
		}
	})

	t.Run("Multiple ExpiredAt changes should always generate new nonces", func(t *testing.T) {
		// Store the current nonces
		firstNonces := make([][]byte, len(pegout.nonces))
		copy(firstNonces, pegout.nonces)

		// Change ExpiredAt again
		unsignedPegout.ExpiredAt = time.Now().Add(2 * time.Hour)
		unsignedPegout.Commitments = map[uint16][]byte{}
		unsignedPegout.CommitmentsMask = make([]byte, 32)

		service.execute(context.Background(), prevDKG)

		// Store the second set of nonces
		secondNonces := make([][]byte, len(pegout.nonces))
		copy(secondNonces, pegout.nonces)

		// Verify second nonces are different from first nonces
		for i, secondNonce := range secondNonces {
			if i < len(firstNonces) && bytes.Equal(secondNonce, firstNonces[i]) {
				t.Fatal("nonces should be different on each ExpiredAt change")
			}
		}

		// Change ExpiredAt a third time
		unsignedPegout.ExpiredAt = time.Now().Add(3 * time.Hour)
		unsignedPegout.Commitments = map[uint16][]byte{}
		unsignedPegout.CommitmentsMask = make([]byte, 32)
		service.execute(context.Background(), prevDKG)

		// Verify third nonces are different from both first and second nonces
		for i, thirdNonce := range pegout.nonces {
			if i < len(firstNonces) && bytes.Equal(thirdNonce, firstNonces[i]) {
				t.Fatal("third nonces should be different from first nonces")
			}
			if i < len(secondNonces) && bytes.Equal(thirdNonce, secondNonces[i]) {
				t.Fatal("third nonces should be different from second nonces")
			}
		}
	})
}
