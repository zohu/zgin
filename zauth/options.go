package zauth

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zutil"
	"time"
)

type Options struct {
	Age                 time.Duration          `yaml:"age" note:"生命周期"`
	AllowMultipleDevice bool                   `yaml:"allow_multiple_device" note:"是否允许多设备同时登陆"`
	AllowIpChange       bool                   `yaml:"allow_ip_change" note:"是否允许IP变化"`
	AllowUaChange       bool                   `yaml:"allow_ua_change" note:"是否允许UA变化"`
	WhiteList           []string               `yaml:"white_list"`
	PathSkip            func(path string) bool `note:"是否跳过校验"`
	Set                 func(ctx context.Context, k, v string, d time.Duration)
	Get                 func(ctx context.Context, k string) string
	Delete              func(ctx context.Context, k string)
}

func (o *Options) Validate() error {
	o.Age = zutil.FirstTruth(o.Age, time.Hour*2)
	if o.Set == nil || o.Get == nil || o.Delete == nil {
		return fmt.Errorf("storage method must be set")
	}
	if o.PathSkip == nil {
		o.PathSkip = func(path string) bool {
			return false
		}
	}
	return validator.New().Struct(o)
}
