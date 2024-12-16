package utils

import "github.com/xssnick/tonutils-go/tvm/cell"

func WriteSlicesToBuffer(slice *cell.Slice) []byte {
	bytes := slice.MustLoadSlice(slice.BitsLeft() / 8)
	for slice.RefsNum() > 0 {
		ref := slice.MustLoadRef()
		bytes = append(bytes, ref.MustLoadSlice(ref.BitsLeft()/8)...)
	}
	return bytes
}
