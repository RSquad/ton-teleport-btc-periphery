package coordinator

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type VSet map[uint16][]byte

func NewVSet(dict *cell.Dictionary) (VSet, error) {
	result, err := parseddict.ParseDict(
		dict,
		func(keySlice *cell.Slice, _ uint) uint16 {
			return uint16(keySlice.MustLoadUInt(16))
		},
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

	return *result, err
}
