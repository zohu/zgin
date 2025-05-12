package zutil

import (
	"github.com/bytedance/sonic"
	"reflect"
)

func FirstTruth[T any](args ...T) T {
	for _, item := range args {
		if !reflect.ValueOf(item).IsZero() {
			return item
		}
	}
	return args[0]
}
func Ptr[T any](in T) *T {
	return &in
}
func Val[T any](in *T) T {
	if in == nil {
		return *new(T)
	}
	return *in
}
func When[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// Clean
// @Description: 清空对象
// @param v
func Clean[T any](v T) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

// AnyToStruct
// @Description: any 转 struct
// @param src
// @param dst
// @return error
func AnyToStruct[T any](src any, dst *T) error {
	b, err := sonic.Marshal(src)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(b, dst)
}
