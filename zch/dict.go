package zch

import (
	"context"
	"fmt"
	"github.com/zohu/zgin/zmap"
	"github.com/zohu/zgin/zutil"
	"strings"
	"time"
)

type DictName string

func (p DictName) String() string {
	return string(p)
}

func (p DictName) Key(args ...string) string {
	if len(args) == 0 {
		return fmt.Sprintf("dict:%s", p)
	}
	return fmt.Sprintf("dict:%s:%s", p, strings.Join(args, ":"))
}

type DictQuery func(ctx context.Context, prefix string) map[string]string
type options struct {
	query  DictQuery
	expire time.Duration
}

var ds = zmap.NewStringer[DictName, *options]()

func NewDict(name DictName, query DictQuery, expire ...time.Duration) {
	expire = append(expire, time.Hour*24)
	ds.Set(name, &options{
		query:  query,
		expire: zutil.FirstTruth(expire...),
	})
}
func Dict(ctx context.Context, name DictName, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is empty")
	}
	if v, err := L().Get(ctx, name.Key(key)); err == nil {
		return v, nil
	}
	if opt, ok := ds.Get(name); ok {
		resp := opt.query(ctx, dictPrefix(key))
		for k, v := range resp {
			_ = L().Set(ctx, name.Key(k), v, opt.expire)
		}
		if len(resp) > 0 {
			return L().Get(ctx, name.Key(key))
		}
	} else {
		return "", fmt.Errorf("dict not found: %s, please call zch.NewDict", name)
	}
	return "", fmt.Errorf("not found: %s", key)
}

func dictPrefix(key string) string {
	arr := []rune(key)
	if len(arr) <= 2 {
		return key
	} else if len(arr) <= 5 {
		return string(arr[:len(arr)-1])
	} else if len(arr) <= 10 {
		return string(arr[:len(arr)-2])
	} else {
		return string(arr[:len(arr)/3*2])
	}
}
