package dict

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	R1Package  [][]byte
	R1Packages map[string]R1Package
)

func parseR1PackageKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadSlice(keySize)

	return fmt.Sprintf("%x", key)
}

func parseR1PackageValue(value *cell.Slice) (R1Package, error) {
	return R1Package(WriteCellsToBuffer(value.MustLoadRef())), nil
}

func NewR1PackageFromDictCell(dictCell *cell.Dictionary) (*R1Packages, error) {
	result, err := parseddict.NewFromDictCell(dictCell, parseR1PackageKey, parseR1PackageValue)
	if err != nil {
		return nil, err
	}
	r1Packages := R1Packages(*result)
	return &r1Packages, nil
}
