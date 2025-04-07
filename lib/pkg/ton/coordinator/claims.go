package coordinator

import (
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGClaimcounters map[uint16]uint16
)

func NewDKGClaimcounters(dict *cell.Dictionary) (DKGClaimcounters, error) {
	claims, err := parseddict.ParseDict(
		dict,
		parseddict.ParseKeyUI16,
		loadUI16Map,
	)

	return *claims, err
}

type DKGClaims struct {
	Mask     *big.Int
	Count    uint16
	Counters DKGClaimcounters
}

func LoadDKGClaims(slice *cell.Slice) (*DKGClaims, error) {
	Mask := slice.MustLoadBigUInt(256)
	Count := uint16(slice.MustLoadUInt(16))

	claimsDict := slice.MustLoadDict(16)
	Counters, err := NewDKGClaimcounters(claimsDict)
	if err != nil {
		return nil, err
	}

	return &DKGClaims{
		Mask, Count, Counters,
	}, nil
}
