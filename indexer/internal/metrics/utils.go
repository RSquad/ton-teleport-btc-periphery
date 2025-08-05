package metrics

import "math"

func MulDivCeil(a, b, c int) int {
	if c == 0 {
		panic("division by zero")
	}
	// Use math.Ceil to round up after floating-point division
	return int(math.Ceil(float64(a*b) / float64(c)))
}
