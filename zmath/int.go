package zmath

import "math"

func Abs(v int64) int64 {
	return int64(math.Abs(float64(v)))
}
