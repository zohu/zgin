package zcpt

import (
	"crypto/md5"
	"fmt"
	"reflect"

	"github.com/bytedance/sonic"
)

func Md5(v any) string {
	if IsNil(v) {
		return ""
	}
	d, _ := sonic.Marshal(v)
	has := md5.Sum(d)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	if rv := reflect.ValueOf(v); rv.Kind() >= reflect.Ptr && rv.Kind() <= reflect.Interface {
		return rv.IsNil()
	}
	return false
}
