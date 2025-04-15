package dkg

import (
	"crypto/rand"
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
		return nil, errors.New("Data is too short; it must be longer than 24 bytes")
	}

	nonce := data[:24]
	dataToDecrypt := data[24:]
	decrypted, ok := box.Open(nil, dataToDecrypt, (*[24]byte)(nonce[:24]), (*[32]byte)(otherPublicKey[:32]), (*[32]byte)(thisPrivateKey[:32]))
	if !ok {
		return nil, fmt.Errorf("Decryption failed for public key '%X'", otherPublicKey)
	}

	return decrypted, nil
}

func DecryptPackages(encryptedPackages map[frost.Identifier]frost.Package, thisPrivateKey []byte, otherPublicKeys map[uint16][]byte) (map[frost.Identifier]frost.Package, error) {
	decryptedPackages := make(map[frost.Identifier]frost.Package)

	for frostIdx, encryptedPackage := range encryptedPackages {
		fromIdx := helpers.FrostToValidatorIdx(frostIdx)
		fromPublicKey, ok := otherPublicKeys[fromIdx]
		if !ok {
			return nil, fmt.Errorf("Public key not found for Oracle {%d}", fromIdx)
		}

		decryptedPackage, err := Decrypt(encryptedPackage.ToBytes(), thisPrivateKey, fromPublicKey)
		if err != nil {
			return nil, fmt.Errorf("Failed to decrypt frost package data from Oracle {%d}. {%w}", fromIdx, err)
		}

		decryptedPackages[frostIdx] = frost.NewPackage(decryptedPackage)
	}

	return decryptedPackages, nil
}
