package utils

import (
	"math"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func BytesPadTo(input []byte, size int) []byte {
	padded := make([]byte, size)
	copy(padded[size-len(input):], input)
	return padded
}

func SplitBytesToCells(buf []byte, splitSize ...int) *cell.Cell {
	size := 1
	if len(splitSize) > 0 {
		size = splitSize[0]
	}

	if size > 127 {
		size = 127
	} else if size < 1 {
		panic("splitSize must be at least 1")
	}

	cellCapacity := int(math.Floor(float64(127/size)) * float64(size))
	cellsCount := int(math.Ceil(float64(len(buf)) / float64(cellCapacity)))

	cells := make([]*cell.Builder, 0, cellsCount)
	offset := 0

	for i := 0; i < cellsCount-1; i++ {
		end := offset + cellCapacity
		if end > len(buf) {
			end = len(buf)
		}
		builder := cell.BeginCell()
		builder.MustStoreBinarySnake(buf[offset:end])
		cells = append(cells, builder)
		offset += cellCapacity
	}

	lastBuilder := cell.BeginCell()
	lastBuilder.MustStoreBinarySnake(buf[offset:])
	cells = append(cells, lastBuilder)

	for i := len(cells) - 1; i >= 1; i-- {
		cells[i-1].MustStoreRef(cells[i].EndCell())
	}

	return cells[0].EndCell()
}
