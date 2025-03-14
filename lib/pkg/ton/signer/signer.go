package signer

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Signer interface {
	SignCell(cell *cell.Cell) []byte
}

type KeySigner struct {
	secret ed25519.PrivateKey
}

func NewKeySigner(secretKey string) Signer {
	secret, err := hex.DecodeString(secretKey)
	if err != nil {
		panic(err)
	}
	return &KeySigner{secret: secret}
}

func (s *KeySigner) SignCell(cell *cell.Cell) []byte {
	return cell.Sign(s.secret)
}
