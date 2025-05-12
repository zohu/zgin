package zutil

import (
	"os"
	"reflect"
	"runtime"
	"strings"
)

func IsDebug() bool {
	return os.Getenv("DEBUG") == "true"
}
func IsLinux() bool {
	return __goos() == "linux"
}
func IsWindows() bool {
	return __goos() == "windows"
}
func IsMac() bool {
	return __goos() == "darwin"
}
func __goos() string {
	return runtime.GOOS
}

func GetFunctionName(i interface{}, seps ...rune) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})
	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}
