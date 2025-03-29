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
	dkgUntilTimestampLoaded  uint64
	dkgUntilTimestampCurrent uint64
	keystoreRootPath         string
	keystore                 *keystore.Keystore
	secret                   ed25519.PrivateKey
}

func NewSessionSigner(keystoreRootPath string) *SessionSigner {
	return &SessionSigner{dkgUntilTimestampLoaded: 0, dkgUntilTimestampCurrent: 0, keystoreRootPath: keystoreRootPath, keystore: nil, secret: nil}
}

func (s *SessionSigner) OnRestartDKG(dkgUntilTimestamp uint64) {
	s.dkgUntilTimestampCurrent = dkgUntilTimestamp
}

func (s *SessionSigner) SignCell(cell *cell.Cell) []byte {
	s.LoadOrGenerateSecret()

	return cell.Sign(s.secret)
}

func (s *SessionSigner) PublicKey() []byte {
	s.LoadOrGenerateSecret()

	data, ok := s.secret.Public().([]byte)
	if !ok {
		panic("Failed to get public key")
	}

	return data
}

func (s *SessionSigner) LoadOrGenerateSecret() {
	// Check cache
	if s.dkgUntilTimestampCurrent > 0 && s.dkgUntilTimestampCurrent == s.dkgUntilTimestampLoaded {
		return
	}

	// Try to load from key storage file
	if s.dkgUntilTimestampCurrent == 0 {
		panic(errors.New("Failed to generate secret, `dkgUntilTimestamp` == 0"))
	}

	if s.keystore == nil {
		keystore, err := keystore.New(s.keystoreRootPath)
		if err != nil {
			panic(fmt.Errorf("Failed to generate secret. %v", err))
		}

		s.keystore = &keystore
	}

	secret := (*s.keystore).LoadSession(s.dkgUntilTimestampCurrent)
	if secret != nil {
		s.secret = secret
		s.dkgUntilTimestampLoaded = s.dkgUntilTimestampCurrent
		return
	}

	// Generate new key pair
	_, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Errorf("Failed to generate secret. %v", err))
	}

	s.secret = secret
	s.dkgUntilTimestampLoaded = s.dkgUntilTimestampCurrent
}
