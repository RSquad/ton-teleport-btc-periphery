package coordinatorcontract

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type DkgPackageParams struct {
	mask  uint64
	count uint64
}

func packageParse(dkg *cell.Slice) (DkgPackageParams, error) {
	mask, _ := dkg.LoadBigUInt(256)
	count, _ := dkg.LoadUInt(16)

	return DkgPackageParams{
		mask:  mask.Uint64(),
		count: count,
	}, nil
}
