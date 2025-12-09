package mutils

import (
	"math"
	"math/big"
)

func MulDivCeil(a, b, c uint) uint {
	if c == 0 {
		panic("division by zero")
	}

	result := (int64(a) * int64(b)) / int64(c)

	if (int64(a)*int64(b))%int64(c) != 0 {
		result++
	}
	if result > math.MaxInt {
		panic("integer overflow")
	}

	return uint(result)
}

func Popcnt(n *big.Int) int {
	count := 0
	zero := big.NewInt(0)
	one := big.NewInt(1)
	temp := new(big.Int)

	for n.Cmp(zero) != 0 {
		count++
		temp.Sub(n, one)
		n.And(n, temp)
	}
	return count
}

func RemovedOneIds(from *big.Int, to *big.Int) []uint16 {
	if from.Sign() < 0 || to.Sign() < 0 {
		return nil
	}

	// bits that were 1 in from and are 0 in to
	diff := new(big.Int).AndNot(from, to)

	var ids []uint16

	for bitIdx, bit := range diff.Bits() {
		for bit != 0 {
			ids = append(ids, uint16(bitIdx))
		}
	}

	return ids
}

func ExtractValuesByIdx(keys []uint16, data map[uint16][]byte) map[uint16][]byte {
	res := make(map[uint16][]byte, len(keys))

	for _, k := range keys {
		if v, ok := data[k]; ok {
			res[k] = v
		}
	}

	return res
}
