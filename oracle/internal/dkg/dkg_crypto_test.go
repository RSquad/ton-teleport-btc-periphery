package dkg

import (
	"crypto/rand"
	"encoding/binary"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"golang.org/x/crypto/nacl/box"
)

func TestDecryptR2Packages_Success(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create test R2 package
	testR2Package := make([]byte, helpers.FrostDkgR2PackageSize)
	rand.Read(testR2Package)

	// Create encrypted packages
	r2packages := map[frost.Identifier]frost.Package{
		helpers.ValidatorIdxToFrost(thisValidatorIdx): frost.NewPackage(testR2Package),
	}

	encryptedPackages, err := EncryptR2Packages(
		r2packages,
		r2PublicKeysX25519,
		privateKey2,
		dkgUntil,
		fromValidatorIdx,
	)
	if err != nil {
		t.Fatalf("Failed to encrypt packages: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: encryptedPackages,
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)
	if err != nil {
		t.Fatalf("DecryptR2Packages failed: %v", err)
	}

	if isCulprit {
		t.Fatalf("Expected no culprit, but got culprit: %d", culprit)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result package, got %d", len(result))
	}

	expectedIdentifier := helpers.ValidatorIdxToFrost(fromValidatorIdx)
	if _, exists := result[expectedIdentifier]; !exists {
		t.Fatalf("Expected package from validator %d not found", fromValidatorIdx)
	}
}

func TestDecryptR2Packages_SkipOwnPackages(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
	}

	// Create packages where FROM validator is same as THIS validator
	packages := map[uint16]map[uint16][]byte{
		thisValidatorIdx: {
			thisValidatorIdx: []byte("dummy_package"),
		},
	}

	// Test DecryptR2Packages - should skip own packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)
	if err != nil {
		t.Fatalf("DecryptR2Packages failed: %v", err)
	}

	if isCulprit {
		t.Fatalf("Expected no culprit, but got culprit: %d", culprit)
	}

	if len(result) != 0 {
		t.Fatalf("Expected 0 result packages (own packages should be skipped), got %d", len(result))
	}
}

func TestDecryptR2Packages_MissingPackageForThisValidator(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)
	otherValidatorIdx := uint16(3)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
	}

	// Create packages where FROM validator doesn't have package for THIS validator
	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			otherValidatorIdx: []byte("dummy_package"), // Package for other validator, not for this one
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for missing package, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_MissingPublicKey(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Don't include public key for fromValidatorIdx
	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		// Missing fromValidatorIdx key
	}

	// Create packages
	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: []byte("dummy_package"),
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for missing public key, but got none")
	}

	if isCulprit {
		t.Fatalf("Expected no culprit (system error), but got culprit: %d", culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when error occurred")
	}
}

func TestDecryptR2Packages_DecryptionFailure(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, _, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create invalid encrypted package (random data that can't be decrypted)
	invalidEncryptedData := make([]byte, 50)
	rand.Read(invalidEncryptedData)

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: invalidEncryptedData,
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for decryption failure, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_DataTooShort(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create data that's too short (less than 8 bytes for DKG until)
	shortData := []byte{1, 2, 3, 4, 5} // Only 5 bytes

	encryptedData, err := Encrypt(shortData, privateKey2, publicKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: encryptedData,
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for data too short, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_WrongDkgUntil(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	wrongDkgUntil := time.Now().Add(2 * time.Hour) // Different timestamp
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create data with wrong DKG until timestamp
	dataToEncrypt := make([]byte, 0)

	// Wrong DKG until timestamp
	tmpBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpBuf, uint64(wrongDkgUntil.Unix()))
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// From validator idx
	tmpBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tmpBuf, fromValidatorIdx)
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// R2 package data
	r2PackageData := make([]byte, helpers.FrostDkgR2PackageSize)
	rand.Read(r2PackageData)
	dataToEncrypt = append(dataToEncrypt, r2PackageData...)

	encryptedData, err := Encrypt(dataToEncrypt, privateKey2, publicKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: encryptedData,
		},
	}

	// Test DecryptR2Packages with correct dkgUntil
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil, // Correct timestamp
	)

	if err == nil {
		t.Fatal("Expected error for wrong DKG until, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_DataTooShortForValidatorIdx(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create data with correct DKG until but missing validator idx
	dataToEncrypt := make([]byte, 0)

	// DKG until timestamp
	tmpBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpBuf, uint64(dkgUntil.Unix()))
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// Missing validator idx (only 1 byte instead of 2)
	dataToEncrypt = append(dataToEncrypt, byte(1))

	encryptedData, err := Encrypt(dataToEncrypt, privateKey2, publicKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: encryptedData,
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for data too short for validator idx, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_WrongFromValidatorIdx(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)
	wrongFromValidatorIdx := uint16(3)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create data with wrong from validator idx
	dataToEncrypt := make([]byte, 0)

	// DKG until timestamp
	tmpBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpBuf, uint64(dkgUntil.Unix()))
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// Wrong from validator idx
	tmpBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tmpBuf, wrongFromValidatorIdx)
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// R2 package data
	r2PackageData := make([]byte, helpers.FrostDkgR2PackageSize)
	rand.Read(r2PackageData)
	dataToEncrypt = append(dataToEncrypt, r2PackageData...)

	encryptedData, err := Encrypt(dataToEncrypt, privateKey2, publicKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: { // This is the actual from validator
			thisValidatorIdx: encryptedData,
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for wrong from validator idx, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_WrongPackageSize(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx := uint16(2)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
		fromValidatorIdx: publicKey2[:],
	}

	// Create data with wrong package size
	dataToEncrypt := make([]byte, 0)

	// DKG until timestamp
	tmpBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpBuf, uint64(dkgUntil.Unix()))
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// From validator idx
	tmpBuf = make([]byte, 2)
	binary.BigEndian.PutUint16(tmpBuf, fromValidatorIdx)
	dataToEncrypt = append(dataToEncrypt, tmpBuf...)

	// Wrong size R2 package data (should be 37 bytes, but we use 20)
	wrongSizeR2PackageData := make([]byte, 20)
	rand.Read(wrongSizeR2PackageData)
	dataToEncrypt = append(dataToEncrypt, wrongSizeR2PackageData...)

	encryptedData, err := Encrypt(dataToEncrypt, privateKey2, publicKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx: {
			thisValidatorIdx: encryptedData,
		},
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)

	if err == nil {
		t.Fatal("Expected error for wrong package size, but got none")
	}

	if !isCulprit {
		t.Fatal("Expected culprit to be found")
	}

	if culprit != fromValidatorIdx {
		t.Fatalf("Expected culprit to be %d, got %d", fromValidatorIdx, culprit)
	}

	if result != nil {
		t.Fatal("Expected nil result when culprit found")
	}
}

func TestDecryptR2Packages_MultipleValidators(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)
	fromValidatorIdx1 := uint16(2)
	fromValidatorIdx2 := uint16(3)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	publicKey2, privateKey2, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	publicKey3, privateKey3, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key 3: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx:  publicKey1[:],
		fromValidatorIdx1: publicKey2[:],
		fromValidatorIdx2: publicKey3[:],
	}

	// Create test R2 packages for both validators
	testR2Package1 := make([]byte, helpers.FrostDkgR2PackageSize)
	rand.Read(testR2Package1)

	testR2Package2 := make([]byte, helpers.FrostDkgR2PackageSize)
	rand.Read(testR2Package2)

	// Create encrypted packages from validator 1
	r2packages1 := map[frost.Identifier]frost.Package{
		helpers.ValidatorIdxToFrost(thisValidatorIdx): frost.NewPackage(testR2Package1),
	}

	encryptedPackages1, err := EncryptR2Packages(
		r2packages1,
		r2PublicKeysX25519,
		privateKey2,
		dkgUntil,
		fromValidatorIdx1,
	)
	if err != nil {
		t.Fatalf("Failed to encrypt packages 1: %v", err)
	}

	// Create encrypted packages from validator 2
	r2packages2 := map[frost.Identifier]frost.Package{
		helpers.ValidatorIdxToFrost(thisValidatorIdx): frost.NewPackage(testR2Package2),
	}

	encryptedPackages2, err := EncryptR2Packages(
		r2packages2,
		r2PublicKeysX25519,
		privateKey3,
		dkgUntil,
		fromValidatorIdx2,
	)
	if err != nil {
		t.Fatalf("Failed to encrypt packages 2: %v", err)
	}

	packages := map[uint16]map[uint16][]byte{
		fromValidatorIdx1: encryptedPackages1,
		fromValidatorIdx2: encryptedPackages2,
	}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)
	if err != nil {
		t.Fatalf("DecryptR2Packages failed: %v", err)
	}

	if isCulprit {
		t.Fatalf("Expected no culprit, but got culprit: %d", culprit)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 result packages, got %d", len(result))
	}

	// Check that both packages are present
	expectedIdentifier1 := helpers.ValidatorIdxToFrost(fromValidatorIdx1)
	expectedIdentifier2 := helpers.ValidatorIdxToFrost(fromValidatorIdx2)

	if _, exists := result[expectedIdentifier1]; !exists {
		t.Fatalf("Expected package from validator %d not found", fromValidatorIdx1)
	}

	if _, exists := result[expectedIdentifier2]; !exists {
		t.Fatalf("Expected package from validator %d not found", fromValidatorIdx2)
	}
}

func TestDecryptR2Packages_EmptyPackages(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
	}

	// Empty packages
	packages := map[uint16]map[uint16][]byte{}

	// Test DecryptR2Packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		packages,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)
	if err != nil {
		t.Fatalf("DecryptR2Packages failed: %v", err)
	}

	if isCulprit {
		t.Fatalf("Expected no culprit, but got culprit: %d", culprit)
	}

	if len(result) != 0 {
		t.Fatalf("Expected 0 result packages, got %d", len(result))
	}
}

func TestDecryptR2Packages_NilPackages(t *testing.T) {
	// Setup test data
	dkgUntil := time.Now().Add(time.Hour)
	thisValidatorIdx := uint16(1)

	// Generate keys
	publicKey1, privateKey1, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	r2PublicKeysX25519 := map[uint16][]byte{
		thisValidatorIdx: publicKey1[:],
	}

	// Test DecryptR2Packages with nil packages
	result, isCulprit, culprit, err := DecryptR2Packages(
		nil,
		thisValidatorIdx,
		r2PublicKeysX25519,
		privateKey1,
		dkgUntil,
	)
	if err != nil {
		t.Fatalf("DecryptR2Packages failed: %v", err)
	}

	if isCulprit {
		t.Fatalf("Expected no culprit, but got culprit: %d", culprit)
	}

	if len(result) != 0 {
		t.Fatalf("Expected 0 result packages, got %d", len(result))
	}
}
