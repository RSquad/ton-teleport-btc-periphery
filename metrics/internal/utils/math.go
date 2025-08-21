package utils

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
