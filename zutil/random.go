package zutil

import (
	"crypto/rand"
	"math"
	"math/big"
	"strconv"
)

func Random(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)-int64(min)))
	return int(n.Int64()) + min
}
func RandomStr(len int) string {
	return strconv.Itoa(Random(int(math.Pow10(len-1)), int(math.Pow10(len)-1)))
}
