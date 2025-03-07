package coordinatorcontract

import (
	"fmt"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGR2 struct {
		mask  *big.Int
		count uint64
		pkgs  *DKGR2Pkgs
	}
	DKGR2Pkgs map[string]*DKGPkgs
)

func parseR2PkgKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadBigUInt(keySize)
	return fmt.Sprintf("%x", key.Bytes())
}

func parseR2PkgValue(valueSlice *cell.Slice) (*DKGPkgs, error) {
	valueSlice.MustLoadUInt(256)

	dict := valueSlice.MustLoadDict(256)
	result, err := NewDKGPkgs(dict)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewR2Pkgs(dict *cell.Dictionary) (*DKGR2Pkgs, error) {
	result, err := parseddict.New(dict, parseR2PkgKey, parseR2PkgValue)
	if err != nil {
		return nil, err
	}
	dkgR2Pkgs := DKGR2Pkgs(*result)
	return &dkgR2Pkgs, nil
}

func NewDKGR2(dict *cell.Dictionary, params *DKGRoundState) (*DKGR2, error) {
	pkgs, err := NewR2Pkgs(dict)
	if err != nil {
		return nil, err
	}
	return &DKGR2{mask: params.mask, count: params.count, pkgs: pkgs}, nil
}

func (d *DKGR2) GetPkgs() *DKGR2Pkgs {
	return d.pkgs
}

func (d *DKGR2Pkgs) GetPkgsByIdentifier(identifier string) map[string][]byte {
	pkgs := (*d)[identifier]
	return pkgs.GetAll()
}
