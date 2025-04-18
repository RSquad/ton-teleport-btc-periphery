package pegoutcontract

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	TxPartsInput struct {
		Amount            *big.Int
		Index             uint32
		Receiver          *address.Address
		BitcoinMerkleRoot []byte
		BitcoinScript     []byte
	}
	TxPartsInputs map[string]*TxPartsInput
	TxInput       struct {
		TxHash []byte
		Data   *TxPartsInput
	}
)

func parseTxPartsInputKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadSlice(keySize)
	return fmt.Sprintf("%x", key)
}

func parseTxPartsInputValue(valueSlice *cell.Slice) (*TxPartsInput, error) {
	if valueSlice == nil {
		return nil, errors.New("value slice is nil")
	}

	amount := valueSlice.MustLoadBigUInt(128)
	index := valueSlice.MustLoadUInt(8)
	merkleRootInt := valueSlice.MustLoadBigUInt(256)

	var bitcoinMerkleRoot []byte
	if merkleRootInt.Sign() != 0 {
		bitcoinMerkleRoot = merkleRootInt.FillBytes(make([]byte, 32))
	}
	receiver := valueSlice.MustLoadAddr()
	bitcoinScriptSlice := valueSlice.MustLoadRef()
	_, bitcoinScript, _ := bitcoinScriptSlice.RestBits()

	return &TxPartsInput{
		Amount:            amount,
		Index:             uint32(index),
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

func (inputs TxPartsInputs) ToSortedSlice() ([]TxInput, error) {
	sortedInputs := make([]TxInput, 0, len(inputs))
	for txHashString := range inputs {
		hash, err := hex.DecodeString(txHashString)
		if err != nil {
			return nil, errors.New("error decoding tx hash " + txHashString + ": " + err.Error())
		}
		sortedInputs = append(sortedInputs, TxInput{
			TxHash: hash,
			Data:   inputs[txHashString],
		})
	}
	slices.SortFunc(sortedInputs, func(a, b TxInput) int {
		return bytes.Compare(a.TxHash, b.TxHash)
	})
	return sortedInputs, nil
}
