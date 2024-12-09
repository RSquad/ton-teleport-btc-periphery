package dict

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func WriteCellsToBuffer(cell *cell.Slice) [][]byte {
	var cellBytes [][]byte

	loadedBytes, _ := cell.LoadSlice(cell.BitsLeft() / 8)
	cellBytes = append(cellBytes, loadedBytes)
	for i := 1; i < cell.RefsNum()+1; i++ {
		cellSlice := cell.MustLoadRef()
		loadedBytes, _ := cellSlice.LoadSlice(cellSlice.BitsLeft() / 8)
		cellBytes = append(cellBytes, loadedBytes)
	}
	return cellBytes
}
