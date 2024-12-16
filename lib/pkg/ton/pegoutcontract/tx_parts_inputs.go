package pegoutcontract

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	TxPartsInput struct {
		Amount            *big.Int
		Index             uint8
		Receiver          *address.Address
		BitcoinMerkleRoot []byte
		BitcoinScript     []byte
	}
	TxPartsInputs map[string]*TxPartsInput
)

func parseTxPartsInputKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadBigUInt(keySize)
	return fmt.Sprintf("%x", key.Bytes())
}

func parseTxPartsInputValue(valueSlice *cell.Slice) (*TxPartsInput, error) {
	if valueSlice == nil {
		return nil, errors.New("value slice is nil")
	}

	amount := valueSlice.MustLoadBigUInt(128)
	index := valueSlice.MustLoadUInt(8)
	bitcoinMerkleRoot := valueSlice.MustLoadSlice(256)
	receiver := valueSlice.MustLoadAddr()
	bitcoinScriptSlice := valueSlice.MustLoadRef()
	_, bitcoinScript, _ := bitcoinScriptSlice.RestBits()

	return &TxPartsInput{
		Amount:            amount,
		Index:             uint8(index),
		BitcoinMerkleRoot: bitcoinMerkleRoot,
		Receiver:          receiver,
		BitcoinScript:     bitcoinScript,
	}, nil
}

func NewTxPartsInputs(dict *cell.Dictionary) (*TxPartsInputs, error) {
	result, err := parseddict.New(dict, parseTxPartsInputKey, parseTxPartsInputValue)
	if err != nil {
		return nil, err
	}
	txPartsInputs := TxPartsInputs(*result)
	return &txPartsInputs, nil
}
