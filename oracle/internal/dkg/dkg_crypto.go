package dkg

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
	"golang.org/x/crypto/nacl/box"
)

func Encrypt(data []byte, thisPrivateKey []byte, otherPublicKey []byte) ([]byte, error) {
	// Generate nonce
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt
	encryptedData := box.Seal(nil, data, &nonce, (*[32]byte)(otherPublicKey[:32]), (*[32]byte)(thisPrivateKey[:32]))

	// Combine nonce and encrypted data
	return append(nonce[:], encryptedData...), nil
}

func Decrypt(data []byte, thisPrivateKey []byte, otherPublicKey []byte) ([]byte, error) {
	// Extract nonce
	if len(data) <= 24 {
		return nil, errors.New("data is too short; it must be longer than 24 bytes")
	}

	nonce := data[:24]
	dataToDecrypt := data[24:]
	decrypted, ok := box.Open(nil, dataToDecrypt, (*[24]byte)(nonce[:24]), (*[32]byte)(otherPublicKey[:32]), (*[32]byte)(thisPrivateKey[:32]))
	if !ok {
		return nil, fmt.Errorf("decryption failed for public key '%X'", otherPublicKey)
	}

	return decrypted, nil
}

func DecryptR2Packages(
	packagesFrom map[uint16][]byte,
	validatorIdx uint16,
	r2PublicKeysX25519 map[uint16][]byte,
	r2PrivateX25519 []byte,
	expectedPackageBatchesCount int) (map[frost.Identifier]frost.Package, bool, uint16 /*culprit*/, error) {

	resultPackages := make(map[frost.Identifier]frost.Package)

	// Search packages to validatorIdx from all other validators
	readOffset := 0
	for fromValidatorIdx, batchedEnctyptedPkgs := range packagesFrom {
		if fromValidatorIdx == validatorIdx {
			continue
		}

		isPackagesFound := false
		for i := 0; i < expectedPackageBatchesCount; i++ {
			if (len(batchedEnctyptedPkgs) - readOffset) < 5 {
				return nil, true, fromValidatorIdx, errors.New("not enough bytes in package (idx and data size)")
			}

			// To validator idx
			toValidatorIdx := binary.BigEndian.Uint16(batchedEnctyptedPkgs[readOffset : readOffset+2])
			readOffset += 2

			// Package size
			packageSize := binary.BigEndian.Uint16(batchedEnctyptedPkgs[readOffset : readOffset+2])
			readOffset += 2

			if (len(batchedEnctyptedPkgs) - readOffset) < int(packageSize) {
				return nil, true, fromValidatorIdx, errors.New("not enough bytes in package (wrong data size)")
			}

			encryptedData := batchedEnctyptedPkgs[readOffset : readOffset+int(packageSize)]
			readOffset += int(packageSize)

			if toValidatorIdx != validatorIdx {
				// Skip package
				continue
			}

			if isPackagesFound {
				return nil, true, fromValidatorIdx, errors.New("duplicate packages to this Oracle")
			}

			isPackagesFound = true

			// Decrypt
			fromPublicKeyX25519, ok := r2PublicKeysX25519[fromValidatorIdx]
			if !ok {
				return nil, false, 0, fmt.Errorf("public key not found for Oracle {%d}", fromValidatorIdx)
			}

			decryptedData, err := Decrypt(encryptedData, r2PrivateX25519, fromPublicKeyX25519)
			if err != nil {
				return nil, true, fromValidatorIdx, fmt.Errorf("failed to decrypt Round2 packages {%d}", fromValidatorIdx)
			}

			frostIdx := helpers.ValidatorIdxToFrost(fromValidatorIdx)
			resultPackages[frostIdx] = frost.NewPackage(decryptedData)
		}
	}

	return resultPackages, false, 0, nil
}
