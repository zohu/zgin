package zutil

import (
	"fmt"
	"strings"
)

// Privacy
// @Description: 私有化数据
// @param str
// @param count 隐藏的个数
// @return string
func Privacy(str string, count int) string {
	if str == "" {
		return ""
	}
	arr := []rune(str)
	l := len(arr)
	if l == 1 {
		return "*"
	}
	if l == 2 {
		return fmt.Sprintf("%s*", string(arr[0]))
	}
	fv := (l - count) / 2
	if fv <= 0 {
		fv = 1
	}
	for i := 0; i < l; i++ {
		if i < fv {
			continue
		}
		if i+1 == l {
			continue
		}
		if i < count+fv {
			arr[i] = '*'
		}
	}
	return string(arr)
}

// PrivacyMust
// @Description:
// @param str
// @param count 隐藏的个数
// @param min 最终的结果最短长度
// @return string
func PrivacyMust(str string, count, min int) string {
	v := Privacy(str, count)
	if v == "" {
		return strings.Repeat("*", min)
	}
	arr := []rune(v)
	l := len(arr)
	if l >= min {
		return v
	}
	pv := (l - count) / 2
	if pv <= 0 {
		pv = 1
	}
	pre := arr[:pv]
	post := arr[pv:]
	return string(pre) + strings.Repeat("*", min-l) + string(post)
}
