package validator

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type OracleSigner struct {
	publicKey  []byte
	privateKey []byte
}

func NewOracleSigner(publicKey []byte, keystore *keystore.Keystore) signer.Signer {
	return &OracleSigner{publicKey: publicKey, privateKey: (*keystore).LoadSecret(publicKey)}
}

func (s *OracleSigner) SignCell(cell *cell.Cell) []byte {
	if s.privateKey == nil {
		return nil
	}

	keySigner := signer.NewKeySigner(hex.EncodeToString(s.privateKey))
	return keySigner.SignCell(cell)
}
