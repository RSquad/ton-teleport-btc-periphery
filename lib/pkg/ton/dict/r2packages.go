package dict

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type R2Packages map[string]*R1Packages

func parseR2PackageKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadSlice(keySize)

	return fmt.Sprintf("%x", key)
}

func parseR2PackageValue(value *cell.Slice) (*R1Packages, error) {
	value.MustLoadUInt(256)

	r1PackageDict := value.MustLoadDict(256)
	r1Packages, err := NewR1PackageFromDictCell(r1PackageDict)
	if err != nil {
		return nil, err
	}

	return r1Packages, nil
}

func NewR2PackageFromDictCell(dictCell *cell.Dictionary) (*R2Packages, error) {
	result, err := parseddict.NewFromDictCell(dictCell, parseR2PackageKey, parseR2PackageValue)
	if err != nil {
		return nil, err
	}
	r2Packages := R2Packages(*result)
	return &r2Packages, nil
}
