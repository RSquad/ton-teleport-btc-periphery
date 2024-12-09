package dict

import (
	"errors"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type VSets map[string][]byte

func parseVSetKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadUInt(keySize)
	return fmt.Sprintf("%x", key)
}

func parseVSetValue(valueSlice *cell.Slice) ([]byte, error) {
	if valueSlice == nil {
		return nil, errors.New("valueSlice is nil")
	}

	tag := valueSlice.MustLoadUInt(8)
	if (tag &^ 0x20) != 0x53 {
		panic("Invalid Validator Descr tag")
	}
	pubKeyTag := valueSlice.MustLoadUInt(32)
	if pubKeyTag != Ed25519PubkeyTag {
		panic("Invalid PublicKey tag")
	}

	return valueSlice.MustLoadSlice(256), nil
}

func NewVSetsFromDictCell(dictCell *cell.Dictionary) (*VSets, error) {
	result, err := parseddict.NewFromDictCell(dictCell, parseVSetKey, parseVSetValue)
	if err != nil {
		return nil, err
	}
	vsets := VSets(*result)
	return &vsets, nil
}
