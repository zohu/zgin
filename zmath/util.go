package zmath

import (
	"github.com/shopspring/decimal"
)

type Number interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func ToInt64(v string) int64 {
	d, _ := decimal.NewFromString(v)
	return d.Round(0).IntPart()
}
func ToFloat64(v string) float64 {
	d, _ := decimal.NewFromString(v)
	return d.InexactFloat64()
}
func ToString(v string) string {
	d, _ := decimal.NewFromString(v)
	return d.String()
}

func Avg[T Number](args ...T) T {
	l := len(args)
	if l == 0 {
		return 0
	}
	sum := NewChain(0)
	for _, n := range args {
		sum = sum.Add(float64(n))
	}
	return T(sum.Div(float64(l)).Float64())
}
