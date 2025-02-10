package pegoutcontract

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TxPartsSignatures map[string][]byte

func parseTxPartsSignatureKey(keySlice *cell.Slice, keySize uint) string {
	return keySlice.MustLoadBigUInt(keySize).String()
}

func parseTxPartsSignatureValue(valueSlice *cell.Slice) ([]byte, error) {
	if valueSlice == nil {
		return nil, errors.New("value slice is nil")
	}

	signatureBytes := valueSlice.MustLoadSlice(valueSlice.BitsLeft())
	return signatureBytes, nil
}

func NewTxPartsSignatures(dict *cell.Dictionary) (*TxPartsSignatures, error) {
	result, err := parseddict.New(dict, parseTxPartsSignatureKey, parseTxPartsSignatureValue)
	if err != nil {
		return nil, err
	}
	txPartsSignatures := TxPartsSignatures(*result)
	return &txPartsSignatures, nil
}
