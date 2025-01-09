package signer

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Signer struct {
	secret ed25519.PrivateKey
}

func New(secretStr string) *Signer {
	secret, err := hex.DecodeString(secretStr)
	if err != nil {
		panic(err)
	}
	return &Signer{secret: secret}
}

func (s *Signer) SignCell(cell *cell.Cell) []byte {
	return cell.Sign(s.secret)
}
