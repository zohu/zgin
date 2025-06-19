package zutil

import (
	"crypto/rand"
	"math/big"
)

func Random(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)-int64(min)))
	return int(n.Int64()) + min
}
func RandomStr(length int) string {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[Random(0, len(charset))]
	}
	return string(b)
}
