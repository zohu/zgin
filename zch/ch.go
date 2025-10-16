package zch

import (
	"context"
	"time"

	"github.com/zohu/zgin/zutil"
	"github.com/zohu/zlog"
)

type L2 struct {
	m *Memory
	r *Redis
}

var l2 *L2

func NewL2(options *Options) *L2 {
	options = zutil.FirstTruth(options, &Options{})
	if err := options.Validate(); err != nil {
		zlog.Fatalf("options is invalid: %v", err)
		return nil
	}
	if l2 == nil {
		l2 = &L2{
			m: NewMemory(options.Expiration, options.CleanInterval, options.Prefix),
			r: NewRedis(options),
		}
	}
	zlog.Infof("zch init success!")
	return l2
}

func L() *L2 {
	return l2
}
func M() *Memory {
	return l2.m
}
func R() *Redis {
	return l2.r
}

func (l *L2) Set(ctx context.Context, k, v string, exp time.Duration) error {
	if err := l.r.Set(ctx, k, v, exp).Err(); err == nil {
		return err
	}
	l.m.Set(k, v, l1(exp))
	return nil
}
func (l *L2) Get(ctx context.Context, k string) (string, error) {
	if v, ok := l.m.Get(k); ok {
		return v, nil
	}
	v, err := l.r.Get(ctx, k).Result()
	if err != nil {
		return "", err
	}
	exp := l.r.ExpireTime(ctx, k).Val()
	l.m.Set(k, v, l1(exp))
	return v, nil
}
func (l *L2) Del(ctx context.Context, k string) error {
	l.m.Delete(k)
	return l.r.Del(ctx, k).Err()
}

func (l *L2) FlushMemory() {
	l.m.Flush()
}

// l1
// @Description: 计算l1缓存的过期时间, l1总是比l2短一些, 且最长是10min，减少内存占用且防止NX虚锁
// @param expiration
// @return time.Duration
func l1(expiration time.Duration) time.Duration {
	return zutil.When(expiration > 10*time.Minute, 10*time.Minute, expiration)
}
