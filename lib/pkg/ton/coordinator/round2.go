package coordinator

import (
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGR2 struct {
		mask         *big.Int
		count        uint64
		PackagesFrom map[uint16][]byte
	}
)

func parseR2PkgValue(valueSlice *cell.Slice) (DKGPkgs, error) {
	valueSlice.MustLoadUInt(256) // skip internal mask

	dict := valueSlice.MustLoadDict(16)
	result, err := NewDKGPkgs(dict)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewR2Pkgs(dict *cell.Dictionary) (map[uint16][]byte, error) {
	result, err := parseddict.ParseDict(dict, parseddict.ParseKeyUI16, readBuffer)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func NewDKGR2(dict *cell.Dictionary, params *DKGRoundState) (*DKGR2, error) {
	pkgs, err := NewR2Pkgs(dict)
	if err != nil {
		return nil, err
	}
	return &DKGR2{mask: params.mask, count: params.count, PackagesFrom: pkgs}, nil
}
