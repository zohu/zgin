package zmath

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strconv"
)

type Chain struct {
	d decimal.Decimal
}

func NewChain(num float64) *Chain {
	return &Chain{d: decimal.NewFromFloat(num)}
}

func (c *Chain) Add(args ...float64) *Chain {
	d := c.d
	for _, arg := range args {
		d = d.Add(decimal.NewFromFloat(arg))
	}
	return &Chain{d}
}
func (c *Chain) Sub(args ...float64) *Chain {
	d := c.d
	for _, arg := range args {
		d = d.Sub(decimal.NewFromFloat(arg))
	}
	return &Chain{d}
}
func (c *Chain) Mul(args ...float64) *Chain {
	d := c.d
	for _, arg := range args {
		d = d.Mul(decimal.NewFromFloat(arg))
	}
	return &Chain{d}
}
func (c *Chain) Div(args ...float64) *Chain {
	d := c.d
	for _, arg := range args {
		d = d.Div(decimal.NewFromFloat(arg))
	}
	return &Chain{d}
}
func (c *Chain) Round(n int32) *Chain {
	return &Chain{c.d.Round(n)}
}
func (c *Chain) Float64() float64 {
	return c.d.InexactFloat64()
}
func (c *Chain) Int64() int64 {
	return c.d.Round(0).IntPart()
}
func (c *Chain) String() string {
	return strconv.FormatFloat(c.Float64(), 'f', -1, 64)
}
func (c *Chain) StringRound(n int32) string {
	return fmt.Sprintf("%f", c.Round(n).Float64())
}
