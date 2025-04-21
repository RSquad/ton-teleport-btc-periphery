package coordinator

import (
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGR2 struct {
		Mask     *big.Int
		Count    uint64
		Packages map[uint16][]byte
	}
)

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
	return &DKGR2{Mask: params.mask, Count: params.count, Packages: pkgs}, nil
}
