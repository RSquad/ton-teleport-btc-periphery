package coordinatorcontract

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGPkgs map[string][]byte
)

func NewDKGPkgs(dict *cell.Dictionary) (*DKGPkgs, error) {
	result, err := parseddict.New(
		dict,
		parseddict.ParseKey,
		func(s *cell.Slice) ([]byte, error) {
			return utils.WriteSlicesToBuffer(s.MustLoadRef()), nil
		},
	)

	return (*DKGPkgs)(result), err
}
