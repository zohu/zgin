package zutil

import (
	"math"
	"strconv"
	"strings"
)

// VersionCompare
// @Description: 对比版本号，版本过旧返回true，否则false
//
//	v1.0.1 > v1.0 = true
//	v1.1 > v1.0.1 = true
//	v1.0.0 > 1.1 = false
//
// @param latest 最低兼容的版本号
// @param local  客户端的版本号
// @return bool
func VersionCompare(latest, local string) bool {
	latest = strings.ReplaceAll(latest, "v", "")
	local = strings.ReplaceAll(local, "v", "")
	arr1 := strings.Split(latest, ".")
	arr2 := strings.Split(local, ".")
	l1 := len(arr1)
	l2 := len(arr2)
	loop := int(math.Max(float64(l1), float64(l2)))
	if l1 < loop {
		for i := l1; i < loop; i++ {
			arr1 = append(arr1, "0")
		}
	}
	if l2 < loop {
		for i := l2; i < loop; i++ {
			arr2 = append(arr2, "0")
		}
	}
	for i := 0; i < loop; i++ {
		p1, _ := strconv.Atoi(arr1[i])
		p2, _ := strconv.Atoi(arr2[i])
		if p1 > p2 {
			return true
		} else if p1 < p2 { // 防止1.0.1 > 1.1
			return false
		}
	}
	return false
}

// VersionCompareForced
// @Description: 判断是否需要强制更新
// @param latest
// @param local
// @return int
func VersionCompareForced(latest, local string) bool {
	latest = strings.ReplaceAll(latest, "v", "")
	local = strings.ReplaceAll(local, "v", "")
	arr1 := strings.Split(latest, ".")
	arr2 := strings.Split(local, ".")
	l1 := len(arr1)
	l2 := len(arr2)
	if l1 < 2 {
		for i := l1; i < 2; i++ {
			arr1 = append(arr1, "0")
		}
	}
	if l2 < 2 {
		for i := l2; i < 2; i++ {
			arr2 = append(arr2, "0")
		}
	}
	for i := 0; i < 2; i++ {
		p1, _ := strconv.Atoi(arr1[i])
		p2, _ := strconv.Atoi(arr2[i])
		if p1 > p2 {
			return true
		} else if p1 < p2 { // 防止1.0 > 2
			return false
		}
	}
	return false
}
