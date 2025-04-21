package dkg

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"golang.org/x/crypto/nacl/box"
)

func Encrypt(data []byte, thisPrivateKey *[32]byte, otherPublicKey *[32]byte) ([]byte, error) {
	// Generate nonce
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt
	encryptedData := box.Seal(nil, data, &nonce, otherPublicKey, thisPrivateKey)

	// Combine nonce and encrypted data
	return append(nonce[:], encryptedData...), nil
}

func Decrypt(data []byte, thisPrivateKey *[32]byte, otherPublicKey *[32]byte) ([]byte, error) {
	// Extract nonce
	if len(data) <= 24 {
		return nil, errors.New("data is too short; it must be longer than 24 bytes")
	}

	nonce := data[:24]
	dataToDecrypt := data[24:]
	decrypted, ok := box.Open(nil, dataToDecrypt, (*[24]byte)(nonce[:24]), otherPublicKey, thisPrivateKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed for public key '%X'", otherPublicKey)
	}

	return decrypted, nil
}

func EncryptR2Packages(
	r2packages map[frost.Identifier]frost.Package,
	r2PublicKeysX25519 map[uint16][]byte,
	r2PrivateX25519 *[32]byte,
) (map[uint16][]byte, error) {
	resultPackages := make(map[uint16][]byte)

	for toIdentificator, r2pkg := range r2packages {
		toValidatorIdx := helpers.FrostToValidatorIdx(toIdentificator)

		// Encrypt r2pkg
		r2PublicKeyX25519, ok := r2PublicKeysX25519[toValidatorIdx]
		if !ok {
			return nil, fmt.Errorf("No X25519 public key was found for Oracle {%d}", toValidatorIdx)
		}

		r2pkgEncrypted, err := Encrypt(
			r2pkg.ToBytes(),
			r2PrivateX25519,
			(*[32]byte)(r2PublicKeyX25519[:]),
		)

		if err != nil {
			return nil, fmt.Errorf("Failed to encrypt R2 packages for Oracle {%d}. %w", toValidatorIdx, err)
		}

		resultPackages[toValidatorIdx] = r2pkgEncrypted
	}

	return resultPackages, nil
}

func DecryptR2Packages(
	packages map[uint16]map[uint16][]byte, // map[FROM]map[TO]data
	thisValidatorIdx uint16,
	r2PublicKeysX25519 map[uint16][]byte,
	r2PrivateX25519 *[32]byte) (map[frost.Identifier]frost.Package /*map[FROM]*/, bool, uint16 /*culprit*/, error) {

	resultPackages := make(map[frost.Identifier]frost.Package)

	// Search packages to thisValidatorIdx from all other validators

	for fromValidatorIdx, toPackages := range packages {
		// Skip our own packages
		if fromValidatorIdx == thisValidatorIdx {
			continue
		}

		// Try to find a package for this validator
		toPackage, ok := toPackages[thisValidatorIdx]

		if !ok {
			return nil, true, fromValidatorIdx, fmt.Errorf("oracle %d didn`t send a package to this oracle %d", fromValidatorIdx, thisValidatorIdx)
		}

		// Decrypt
		fromPublicKeyX25519, ok := r2PublicKeysX25519[fromValidatorIdx]
		if !ok {
			return nil, false, 0, fmt.Errorf("public key not found for Oracle {%d}", fromValidatorIdx)
		}

		decryptedData, err := Decrypt(toPackage, r2PrivateX25519, (*[32]byte)(fromPublicKeyX25519[:32]))
		if err != nil {
			return nil, true, fromValidatorIdx, fmt.Errorf("failed to decrypt Round2 packages {%d}", fromValidatorIdx)
		}

		frostIdx := helpers.ValidatorIdxToFrost(fromValidatorIdx)
		resultPackages[frostIdx] = frost.NewPackage(decryptedData)
	}

	return resultPackages, false, 0, nil
}
