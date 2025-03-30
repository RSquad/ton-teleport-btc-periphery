package validator

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"

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

func (s *SessionSigner) OnNewDKG(dkgUntilTimestamp int64) {
	// Verify
	if dkgUntilTimestamp == 0 {
		panic(errors.New("failed to generate secret, `dkgUntilTimestamp` == 0"))
	}

	if s.dkgUntilTimestamp == dkgUntilTimestamp {
		return
	}

	// Try to load from key storage file
	secret := s.keystore.LoadSessionTS(dkgUntilTimestamp)
	if secret != nil {
		s.secret = secret
		s.dkgUntilTimestamp = dkgUntilTimestamp
		return
	}

	// Generate new key pair
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Errorf("failed to generate secret. %v", err))
	}

	s.secret = secret
	s.dkgUntilTimestamp = dkgUntilTimestamp

	// Save to file (with dkgUntilTimestamp name)
	err = s.keystore.StoreSessionTS(dkgUntilTimestamp, s.secret)
	if err != nil {
		panic(fmt.Errorf("failed to save secret. %v", err))
	}

	// Save to file (with PublicKey name)
	err = s.keystore.StoreSessionPubKey(s.PublicKey(), s.secret)
	if err != nil {
		panic(fmt.Errorf("failed to save secret. %v", err))
	}
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

func (s *SessionSigner) TryLoadFromFile(publicKey []byte) bool {
	secret := s.keystore.LoadSessionPubKey(publicKey)
	if secret != nil {
		s.secret = secret
		return true
	}

	return false
}
