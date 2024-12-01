package pegoutcontract

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TxPartsSignatures map[string][]byte

func parseTxPartsSignatureKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadBigUInt(keySize)
	return key.String()
}

func parseTxPartsSignatureValue(valueSlice *cell.Slice) ([]byte, error) {
	if valueSlice == nil {
		return nil, errors.New("valueSlice is nil")
	}

	signatureBytes := valueSlice.MustLoadSlice(64 * 8)
	return signatureBytes, nil
}

func NewTxPartsSignaturesFromDictCell(dictCell *cell.Dictionary) (*TxPartsSignatures, error) {
	result, err := parseddict.NewFromDictCell(dictCell, parseTxPartsSignatureKey, parseTxPartsSignatureValue)
	if err != nil {
		return nil, err
	}
	txPartsSignatures := TxPartsSignatures(*result)
	return &txPartsSignatures, nil
}
