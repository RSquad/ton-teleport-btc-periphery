package coordinator

import (
	"math/big"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type DKGR1 struct {
	Mask     *big.Int
	Count    uint64
	Packages DKGPkgs
}

func NewDKGR1(d *cell.Dictionary, p *DKGRoundState) (*DKGR1, error) {
	pkg, err := NewDKGPkgs(d)
	return &DKGR1{Mask: p.mask, Count: p.count, Packages: pkg}, err
}
