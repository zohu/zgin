package zmath

import "github.com/shopspring/decimal"

func Add(d1, d2 float64) float64 {
	return decimal.NewFromFloat(d1).Add(decimal.NewFromFloat(d2)).InexactFloat64()
}
func Sub(f1, f2 float64) float64 {
	return decimal.NewFromFloat(f1).Sub(decimal.NewFromFloat(f2)).InexactFloat64()
}
func Mul(f1, f2 float64) float64 {
	return decimal.NewFromFloat(f1).Mul(decimal.NewFromFloat(f2)).InexactFloat64()
}
func Div(f1, f2 float64) float64 {
	if f2 == 0 {
		return 0
	}
	return decimal.NewFromFloat(f1).Div(decimal.NewFromFloat(f2)).InexactFloat64()
}
func Round(f float64, n int32) float64 {
	return decimal.NewFromFloat(f).Round(n).InexactFloat64()
}
func RoundInt64(f float64) int64 {
	return decimal.NewFromFloat(f).Round(0).IntPart()
}
