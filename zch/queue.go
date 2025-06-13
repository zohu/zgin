package zch

import (
	"context"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/redis/go-redis/v9"
	"github.com/zohu/zgin/zlog"
	"time"
)

type Topic struct {
	prefix Prefix
}

var process = make(map[string]bool)

func NewTopic(prefix Prefix) *Topic {
	t := &Topic{
		prefix: prefix,
	}
	if _, ok := process[t.prefix.Key()]; !ok {
		go t.processDelayed(t.prefix)
		process[t.prefix.Key()] = true
	}
	return t
}

func (t *Topic) Publish(ctx context.Context, message string, delay ...time.Duration) error {
	if len(delay) > 0 {
		return R().ZAdd(ctx, t.prefix.Key(), redis.Z{
			Score:  float64(time.Now().Unix() + int64(delay[0].Seconds())),
			Member: message,
		}).Err()
	}
	return R().LPush(ctx, t.prefix.Key(), message).Err()
}

// Subscribe
// @Description: 订阅消息，handler返回err时，消息不消费
// @receiver t
// @param ctx
// @param handler
func (t *Topic) Subscribe(ctx context.Context, handler func(string) error) {
	for {
		result, err := R().BRPop(ctx, 0, t.prefix.Key()).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			zlog.Warnf("zch queue subscribe err: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}
		err = retry.Do(
			func() error {
				return handler(result[1])
			},
			retry.Attempts(3),
			retry.Delay(time.Millisecond*100),
			retry.DelayType(retry.FixedDelay),
			retry.LastErrorOnly(true),
		)
		if err != nil {
			zlog.Warnf("zch queue subscribe handler err: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}
		R().RPop(ctx, t.prefix.Key())
	}
}

func (t *Topic) processDelayed(p Prefix) {
	ctx := context.Background()
	for {
		now := time.Now().Unix()
		entries := R().ZRangeByScoreWithScores(ctx, p.Key(), &redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", now),
		}).Val()
		for _, entry := range entries {
			R().LPush(ctx, p.Key(), entry.Member)
			R().ZRem(ctx, p.Key(), entry.Member)
		}
		time.Sleep(1 * time.Second)
	}
}
