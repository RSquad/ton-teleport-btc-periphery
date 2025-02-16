package signer

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const TestSecret = ""

type Signer struct {
	secret ed25519.PrivateKey
}

func New() *Signer {
	secret, err := hex.DecodeString(TestSecret)
	if err != nil {
		panic(err)
	}
	return &Signer{secret: secret}
}

func (s *Signer) SignCell(cell *cell.Cell) []byte {
	return cell.Sign(s.secret)
}
