package zws

import (
	"github.com/zohu/zgin/zutil"
	"net/http"
	"time"
)

type Options struct {
	// 服务端配置
	// 写超时，默认10s
	WriteWait time.Duration `yaml:"write_wait"`
	// 支持接受的消息最大长度，默认2048
	MaxMessageSize int64 `yaml:"max_message_size"`
	// 消息发送缓冲池大小，默认2048，缓冲区满时，会发送失败并返回错误
	MessageBufferSize int `yaml:"message_buffer_size"`
	// 是否开启Gzip压缩，默认关闭，如果开启，客户端也要相应处理
	Gzip bool `yaml:"gzip"`
	// 重连间隔，默认 1s
	BackoffDuration time.Duration `yaml:"backoff_duration"`
	// 重连次数，默认 60
	BackoffTimes uint `yaml:"backoff_times"`

	// 房间配置，房间最大容量，默认100w
	HomeMaxSize int `yaml:"home_max_size"`
	// 广播时，最大线程数，默认100
	HomeBroadcastPoolMaxSize int64 `yaml:"home_broadcast_pool_max_size"`

	// 服务配置
	ServeCheckOrigin    func(r *http.Request) bool
	ServeAllowAllOrigin bool
}

func (o *Options) Validate() {
	o.WriteWait = zutil.FirstTruth(o.WriteWait, 10*time.Second)
	o.MaxMessageSize = zutil.FirstTruth(o.MaxMessageSize, 2048)
	o.MessageBufferSize = zutil.FirstTruth(o.MessageBufferSize, 2048)
	o.BackoffDuration = zutil.FirstTruth(o.BackoffDuration, time.Second)
	o.BackoffTimes = zutil.FirstTruth(o.BackoffTimes, 60)
	o.HomeMaxSize = zutil.FirstTruth(o.HomeMaxSize, 1000000)
	o.HomeBroadcastPoolMaxSize = zutil.FirstTruth(o.HomeBroadcastPoolMaxSize, 100)
}
