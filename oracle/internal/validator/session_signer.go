package validator

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type SessionSigner struct {
	keystore          keystore.Keystore
	secret            ed25519.PrivateKey
	dkgUntilTimestamp int64
}

func NewSessionSigner(keystore keystore.Keystore) *SessionSigner {
	return &SessionSigner{keystore: keystore, secret: nil, dkgUntilTimestamp: 0}
}

func (s *SessionSigner) OnNewDKG(dkgUntilTimestamp int64) error {
	// Verify
	if dkgUntilTimestamp == 0 {
		return errors.New("failed to generate secret, `dkgUntilTimestamp` == 0")
	}

	if s.dkgUntilTimestamp == dkgUntilTimestamp {
		return nil
	}

	// Try to load from key storage file
	logger.Log.Info().Msgf("Try to find session keypair for DKG (until %d)", dkgUntilTimestamp)
	secret := s.keystore.LoadSessionTS(dkgUntilTimestamp)
	if secret != nil {
		s.secret = secret
		s.dkgUntilTimestamp = dkgUntilTimestamp
		logger.Log.Info().Msgf("Session keypair for DKG (until %d) was loaded from file", dkgUntilTimestamp)
		return nil
	}

	// Generate new key pair
	logger.Log.Info().Msgf("Generating new keypair for DKG (until %d)", dkgUntilTimestamp)
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate new keypair for DKG (until %d). %v", dkgUntilTimestamp, err)
	}

	s.secret = secret
	s.dkgUntilTimestamp = dkgUntilTimestamp

	// Save to file (with dkgUntilTimestamp name)
	err = s.keystore.StoreSessionTS(dkgUntilTimestamp, s.secret)
	if err != nil {
		return fmt.Errorf("failed to save keypair. %v", err)
	}

	return nil
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

/*
		// Search in session
		for idx, pubkey := range vset {
			sessionPublicKey := v.sessionSigner.PublicKey()
			if bytes.Equal(pubkey, sessionPublicKey) {
				return &KeyInfo{
					KeyID:     sessionPublicKey,
					VsetIdx:   idx,
					PublicKey: sessionPublicKey,
				}, nil
			}
		}

	// Try searching in the previous session keys
	for idx, pubkey := range vset {
		ok := v.sessionSigner.TryLoadFromFile(pubkey)
		if ok {
			return &KeyInfo{
				KeyID:     pubkey,
				VsetIdx:   idx,
				PublicKey: pubkey,
			}, nil
		}
	}
*/
