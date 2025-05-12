package zauth

import (
	"context"
	"fmt"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"time"
)

type Options struct {
	Prefix              string                 `note:"前缀"`
	Age                 time.Duration          `note:"生命周期"`
	AllowMultipleDevice bool                   `note:"是否允许多设备同时登陆"`
	AllowIpChange       bool                   `note:"是否允许IP变化"`
	AllowUaChange       bool                   `note:"是否允许UA变化"`
	PathSkip            func(path string) bool `note:"是否跳过校验"`
	Set                 func(ctx context.Context, k, v string, d time.Duration)
	Expire              func(ctx context.Context, k string, d time.Duration)
	Get                 func(ctx context.Context, k string) string
	Delete              func(ctx context.Context, k string)
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(opts *Options) {
		opts.Prefix = prefix
	}
}
func WithAge(age time.Duration) Option {
	return func(opts *Options) {
		opts.Age = age
	}
}
func WithAllowMultipleDevice(multipleDevice bool) Option {
	return func(opts *Options) {
		opts.AllowMultipleDevice = multipleDevice
	}
}
func WithAllowIpChange(allowIpChange bool) Option {
	return func(opts *Options) {
		opts.AllowIpChange = allowIpChange
	}
}
func WithAllowUaChange(allowUaChange bool) Option {
	return func(opts *Options) {
		opts.AllowUaChange = allowUaChange
	}
}
func WithPathSkip(pathSkip func(path string) bool) Option {
	return func(opts *Options) {
		opts.PathSkip = pathSkip
	}
}
func WithStorage(
	set func(ctx context.Context, k, v string, d time.Duration),
	expire func(ctx context.Context, k string, d time.Duration),
	get func(ctx context.Context, k string) string,
	delete func(ctx context.Context, k string),
) Option {
	return func(opts *Options) {
		opts.Set = set
		opts.Expire = expire
		opts.Get = get
		opts.Delete = delete
	}
}

func (o *Options) Validate() {
	o.Prefix = zutil.FirstTruth(o.Prefix, "zauth")
	o.Age = zutil.FirstTruth(o.Age, time.Hour*2)
	if o.Set == nil || o.Expire == nil || o.Get == nil || o.Delete == nil {
		zlog.Fatalf("storage method must be set")
	}
}
func (o *Options) WithPrefix(k string) string {
	return fmt.Sprintf("%s:%s", o.Prefix, k)
}
