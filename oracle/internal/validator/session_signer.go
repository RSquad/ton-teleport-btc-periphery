package validator

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type SessionSigner struct {
	secret ed25519.PrivateKey
}

func NewSessionSigner(keystore keystore.Keystore, dkgUntilTimestamp int64) (*SessionSigner, error) {
	// Just in case, we have already created a session `dkgUntilTimestamp`
	sessionSigner, err := LoadSessionSigner(keystore, dkgUntilTimestamp)
	if err == nil {
		return sessionSigner, nil
	}

	// Generate new key pair
	logger.Log.Info().Msgf("Generating new keypair for DKG (until %d)", dkgUntilTimestamp)
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new keypair for DKG (until %d). %v", dkgUntilTimestamp, err)
	}

	// Save to file (with dkgUntilTimestamp name)
	err = keystore.StoreSession(dkgUntilTimestamp, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to save keypair. %v", err)
	}

	return &SessionSigner{secret}, nil
}

func LoadSessionSigner(keystore keystore.Keystore, dkgUntilTimestamp int64) (*SessionSigner, error) {
	// Try to load from key storage file
	logger.Log.Info().Msgf("Try to load session keypair for DKG (until %d)", dkgUntilTimestamp)
	secret := keystore.LoadSession(dkgUntilTimestamp)
	if secret == nil {
		return nil, fmt.Errorf("failed to load session keypair for DKG (until %d)", dkgUntilTimestamp)
	}

	logger.Log.Info().Msgf("Session keypair for DKG (until %d) was loaded from file", dkgUntilTimestamp)
	return &SessionSigner{secret}, nil
}

func (s *SessionSigner) SignCell(cell *cell.Cell) []byte {
	return cell.Sign(s.secret)
}

func (s *SessionSigner) PublicKey() []byte {
	data, ok := s.secret.Public().(ed25519.PublicKey)
	if !ok {
		panic("failed to get public key")
	}

	return data
}
