package coordinator

import (
	"errors"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGClaimcounters map[string][]byte
)

func NewDKGClaimcounters(dict *cell.Dictionary) (DKGClaimcounters, error) {
	claims, err := parseddict.New(
		dict,
		parseddict.ParseKey,
		func(s *cell.Slice) ([]byte, error) {
			return utils.WriteSlicesToBuffer(s.MustLoadRef()), nil
		},
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

	claimsSlice := slice.MustLoadMaybeRef()
	if claimsSlice != nil {
		return nil, errors.New("claimsSlice is nil")
	}

	claimsDict := slice.MustLoadDict(16)
	Counters, err := NewDKGClaimcounters(claimsDict)
	if err != nil {
		return nil, err
	}

	return &DKGClaims{
		Mask, Count, Counters,
	}, nil
}
