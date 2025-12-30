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
	component := "SessionSigner"

	// Just in case, we have already created a session `dkgUntilTimestamp`
	sessionSigner, err := LoadSessionSigner(keystore, dkgUntilTimestamp)
	if err == nil {
		logger.Log.Info().
			Str("component", component).
			Int64("dkg_until_timestamp", dkgUntilTimestamp).
			Msg("Loaded existing session signer")
		return sessionSigner, nil
	}

	// Generate new key pair
	logger.Log.Info().
		Str("component", component).
		Int64("dkg_until_timestamp", dkgUntilTimestamp).
		Msg("Generating new keypair for DKG")

	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new keypair for DKG (until %d). %v", dkgUntilTimestamp, err)
	}

	// Save to file (with dkgUntilTimestamp name)
	err = keystore.StoreSession(dkgUntilTimestamp, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to save keypair. %v", err)
	}

	logger.Log.Info().
		Str("component", component).
		Int64("dkg_until_timestamp", dkgUntilTimestamp).
		Msg("Generated and stored new session keypair")

	return &SessionSigner{secret}, nil
}

func LoadSessionSigner(keystore keystore.Keystore, dkgUntilTimestamp int64) (*SessionSigner, error) {
	component := "SessionSigner"

	logger.Log.Debug().
		Str("component", component).
		Int64("dkg_until_timestamp", dkgUntilTimestamp).
		Msg("Attempting to load session keypair")

	// Try to load from key storage file
	secret := keystore.LoadSession(dkgUntilTimestamp)
	if secret == nil {
		return nil, fmt.Errorf("failed to load session keypair for DKG (until %d)", dkgUntilTimestamp)
	}

	logger.Log.Info().
		Str("component", component).
		Int64("dkg_until_timestamp", dkgUntilTimestamp).
		Msg("Session keypair loaded from storage")

	return &SessionSigner{secret}, nil
}

func (s *SessionSigner) SignCell(cell *cell.Cell) []byte {
	component := "SessionSigner"

	signature := cell.Sign(s.secret)

	logger.Log.Debug().
		Str("component", component).
		Int("signature_length", len(signature)).
		Msg("Cell signed successfully")

	return signature
}

func (s *SessionSigner) PublicKey() []byte {
	component := "SessionSigner"

	data, ok := s.secret.Public().(ed25519.PublicKey)
	if !ok {
		panic("failed to get public key")
	}

	logger.Log.Debug().
		Str("component", component).
		Int("public_key_length", len(data)).
		Msg("Public key retrieved")

	return data
}
