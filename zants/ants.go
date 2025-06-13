package zants

import (
	"github.com/go-playground/validator/v10"
	"github.com/panjf2000/ants/v2"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
)

type Options struct {
	MultiSize int `yaml:"multi_size" note:"池子总量"`
	PoolSize  int `yaml:"pool_size" note:"每个池子的大小"`
}

func (o *Options) Validate() error {
	o.MultiSize = zutil.FirstTruth(o.MultiSize, 1)
	o.PoolSize = zutil.FirstTruth(o.PoolSize, 100)
	return validator.New().Struct(o)
}

type PoolStatus struct {
	Cap     int32 `json:"cap" note:"容量"`
	Running int32 `json:"running" note:"运行中"`
	Waiting int32 `json:"waiting" note:"等待中"`
	Idle    int32 `json:"idle" note:"空闲中"`
}

var multiPool *ants.MultiPool

func New(options *Options) {
	options = zutil.FirstTruth(options, &Options{})
	if err := options.Validate(); err != nil {
		zlog.Fatalf("options is invalid: %v", err)
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
