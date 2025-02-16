package coordinatorcontract

import (
	"math/big"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type DKGR1 struct {
	mask  *big.Int
	count uint64
	pkgs  *DKGPkgs
}

func NewDKGR1(d *cell.Dictionary, p *DKGRoundState) (*DKGR1, error) {
	pkg, err := NewDKGPkgs(d)
	return &DKGR1{mask: p.mask, count: p.count, pkgs: pkg}, err
}
