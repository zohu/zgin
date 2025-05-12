package zants

import (
	"github.com/panjf2000/ants/v2"
	"github.com/zohu/zgin/zlog"
)

type Options struct {
	MultiSize int
	PoolSize  int
}
type Option func(*Options)

// WithMultiSize
// @Description: 线程池数量
// @param multiSize
// @return Option
func WithMultiSize(multiSize int) Option {
	return func(opts *Options) {
		opts.MultiSize = multiSize
	}
}

// WithPoolSize
// @Description: 每池线程数量
// @param poolSize
// @return Option
func WithPoolSize(poolSize int) Option {
	return func(opts *Options) {
		opts.PoolSize = poolSize
	}
}

type PoolStatus struct {
	Cap     int32 `json:"cap" note:"容量"`
	Running int32 `json:"running" note:"运行中"`
	Waiting int32 `json:"waiting" note:"等待中"`
	Idle    int32 `json:"idle" note:"空闲中"`
}

var multiPool *ants.MultiPool

func New(opts ...Option) {
	options := &Options{
		MultiSize: 1,
		PoolSize:  10,
	}
	for _, opt := range opts {
		opt(options)
	}
	p, err := ants.NewMultiPool(options.MultiSize, options.PoolSize, ants.LeastTasks, ants.WithLogger(zlog.NewZLogger(nil)))
	if err != nil {
		zlog.Fatalf("new multi pool error: %v", err)
	}
	multiPool = p
	zlog.Infof("init zants success, size=%dx%d", options.MultiSize, options.PoolSize)
}

// Submit
// @Description: 提交一个函数到池中执行
// @param fn
func Submit(fn func()) {
	if err := multiPool.Submit(fn); err != nil {
		zlog.Errorf("submit fn error: %v", err)
	}
}

// Status
// @Description: 获取池状态
// @return *PoolStatus
func Status() *PoolStatus {
	return &PoolStatus{
		Cap:     int32(multiPool.Cap()),
		Running: int32(multiPool.Running()),
		Waiting: int32(multiPool.Waiting()),
		Idle:    int32(multiPool.Free()),
	}
}

// Tune
// @Description: 调整每个池大小
// @param size
func Tune(size int) {
	multiPool.Tune(size)
}
