package coordinator

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	DKGPkgs map[uint16][]byte
)

func NewDKGPkgs(dict *cell.Dictionary) (DKGPkgs, error) {
	pkgs, err := parseddict.ParseDict(
		dict,
		parseddict.ParseKeyUI16,
		func(s *cell.Slice) ([]byte, error) {
			return utils.WriteSlicesToBuffer(s.MustLoadRef()), nil
		},
	)

	return *pkgs, err
}
