package coordinator

import (
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type PubkeyData struct {
	PubkeyPackage []byte
	InternalKey   []byte
}

type DKGR3 struct {
	Mask  *big.Int
	Count uint16
	Data  *PubkeyData
}

func LoadDKGR3(slice *cell.Slice) (*DKGR3, error) {
	mask := slice.MustLoadBigUInt(256)
	count := uint16(slice.MustLoadUInt(16))

	pkgSlice := slice.MustLoadMaybeRef()

	var data *PubkeyData
	if pkgSlice != nil {
		data = &PubkeyData{
			PubkeyPackage: utils.WriteSlicesToBuffer(pkgSlice.MustLoadRef()),
			InternalKey:   pkgSlice.MustLoadSlice(256),
		}
	}

	return &DKGR3{
		Mask:  mask,
		Count: count,
		Data:  data,
	}, nil
}
