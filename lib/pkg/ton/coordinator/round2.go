package coordinator

import (
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGR2 struct {
		mask     *big.Int
		count    uint64
		Packages DKGR2Pkgs
	}
	DKGR2Pkgs map[uint16]DKGPkgs
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

func NewR2Pkgs(dict *cell.Dictionary) (DKGR2Pkgs, error) {
	result, err := parseddict.ParseDict(dict, parseddict.ParseKeyUI16, parseR2PkgValue)
	if err != nil {
		return nil, err
	}
	dkgR2Pkgs := DKGR2Pkgs(*result)
	return dkgR2Pkgs, nil
}

func NewDKGR2(dict *cell.Dictionary, params *DKGRoundState) (*DKGR2, error) {
	pkgs, err := NewR2Pkgs(dict)
	if err != nil {
		return nil, err
	}
	return &DKGR2{mask: params.mask, count: params.count, Packages: pkgs}, nil
}
