package metrics

import "math"

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
