package validator

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type ValidatorSigner struct {
	validatorConsole *ValidatorConsole
	publicKeyID      string
}

func NewValidatorSigner(validatorConsole *ValidatorConsole, publicKeyID string) signer.Signer {
	return &ValidatorSigner{validatorConsole: validatorConsole, publicKeyID: publicKeyID}
}

func (s *ValidatorSigner) SignCell(cell *cell.Cell) []byte {
	cellHex := hex.EncodeToString(cell.Hash())
	signature, err := s.validatorConsole.Sign(s.publicKeyID, cellHex)
	if err != nil {
		return nil
	}

	return signature
}

func (s *ValidatorSigner) PublicKey() []byte {
	data, err := hex.DecodeString(s.publicKeyID)
	if err != nil {
		panic(err)
	}

	return data
}
