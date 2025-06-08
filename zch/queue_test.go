package zch

import (
	"context"
	"github.com/zohu/zgin/zlog"
	"testing"
	"time"
)

func TestTopic(t *testing.T) {
	NewL2(&Options{
		Expiration:    time.Minute,
		CleanInterval: time.Minute,
		Addrs:         []string{"localhost:8011"},
		Database:      0,
		Password:      "JCFkQYex4f",
		Prefix:        "cs",
	})
	var prefix Prefix = "test_queue"
	topic := NewTopic(prefix)

	ctx := context.Background()
	go topic.Subscribe(ctx, func(msg string) error {
		zlog.Infof("receive msg: %s", msg)
		return nil
	})
	zlog.Infof("publish msg")
	_ = topic.Publish(ctx, "hello world")
	zlog.Infof("publish msg after 5s")
	_ = topic.Publish(ctx, "hello world after 5s", time.Second*5)
	zlog.Infof("publish msg after 8s")
	_ = topic.Publish(ctx, "hello world after 8s", time.Second*8)

	time.Sleep(time.Second * 20)
}
