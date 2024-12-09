package dict

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	SigningSharesKey   [32]byte
	SigningSharesValue *cell.Cell
)
type SigningShares map[string]*cell.Cell

func parseSigningShareKey(keySlice *cell.Slice, keySize uint) string {
	key := keySlice.MustLoadSlice(keySize)

	return fmt.Sprintf("%x", key)
}

func parseSigningShareValue(value *cell.Slice) (*cell.Cell, error) {
	return value.MustToCell(), nil
}

func NewSigningSharesFromDictCell(dictCell *cell.Dictionary) (*SigningShares, error) {
	result, err := parseddict.NewFromDictCell(dictCell, parseSigningShareKey, parseSigningShareValue)
	if err != nil {
		return nil, err
	}
	signingShares := SigningShares(*result)
	return &signingShares, nil
}
