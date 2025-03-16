package utils

import "github.com/xssnick/tonutils-go/tvm/cell"

func WriteSlicesToBuffer(slice *cell.Slice) []byte {
	bytes := slice.MustLoadSlice(slice.BitsLeft())
	for slice.RefsNum() > 0 {
		slice = slice.MustLoadRef()
		bytes = append(bytes, slice.MustLoadSlice(slice.BitsLeft())...)
	}
	return bytes
}
