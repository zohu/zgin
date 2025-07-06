package zch

import (
	"fmt"
	"strings"
)

type Prefix string

func (p Prefix) Key(args ...string) string {
	if len(args) == 0 {
		return string(p)
	}
	return fmt.Sprintf("%s:%s", string(p), strings.Join(args, ":"))
}

// 系统预留前缀
const (
	PrefixI18n       Prefix = "z18n"
	PrefixAuthPreID  Prefix = "auth:pre"
	PrefixAuthToken  Prefix = "auth:user"
	PrefixAuthAction Prefix = "auth:action"
)
