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
	Mask := slice.MustLoadBigUInt(256)
	Count := uint16(slice.MustLoadUInt(16))

	pkgSlice := slice.MustLoadMaybeRef()
	var pubkeyPackage []byte
	if pkgSlice != nil {
		pubkeyPackage = utils.WriteSlicesToBuffer(pkgSlice)
	}

	Data := &PubkeyData{
		PubkeyPackage: pubkeyPackage,
		InternalKey:   slice.MustLoadSlice(32),
	}

	return &DKGR3{
		Mask, Count, Data,
	}, nil
}
