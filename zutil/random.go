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

type WeightedNode[T any] struct {
	Value  T
	Weight int
}

// RandomWeighted
// @Description: 加权随机
// @param nodes
// @return T
func RandomWeighted[T any](nodes []WeightedNode[T]) T {
	if len(nodes) == 0 {
		return *new(T)
	}
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += node.Weight
	}
	r := Random(1, totalWeight)
	currentWeight := 0
	for _, node := range nodes {
		currentWeight += node.Weight
		if r <= currentWeight {
			return node.Value
		}
	}
	return nodes[0].Value
}
