package coordinatorcontract

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type VSet map[string][]byte

func NewVSet(dict *cell.Dictionary) (*VSet, error) {
	result, err := parseddict.New(
		dict,
		parseddict.ParseKey,
		func(valueSlice *cell.Slice) ([]byte, error) {
			tag, err := valueSlice.LoadUInt(8)
			if err != nil {
				return nil, err
			}
			if (tag &^ 0x20) != 0x53 {
				return nil, errors.New("invalid validator descr tag")
			}
			pubKeyTag, err := valueSlice.LoadUInt(32)
			if err != nil {
				return nil, err
			}
			if pubKeyTag != Ed25519PubkeyTag {
				return nil, errors.New("invalid public key tag")
			}

			return valueSlice.LoadSlice(256)
		},
	)

	return (*VSet)(result), err
}
