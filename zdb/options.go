package zdb

import (
	"fmt"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"time"
)

type Options struct {
	Host              string        `validate:"required" note:"数据库地址"`
	Port              int           `validate:"required" note:"数据库端口"`
	User              string        `validate:"required" note:"数据库用户"`
	Pass              string        `validate:"required" note:"数据库密码"`
	DB                string        `validate:"required" note:"数据库名"`
	Config            string        `note:"数据库配置"`
	MaxIdle           int           `note:"最大闲置连接数"`
	MaxAlive          int           `note:"最大存活连接数"`
	MaxAliveLife      time.Duration `note:"最大存活时间"`
	LogSlow           time.Duration `note:"慢阈值，秒"`
	LogIgnoreNotFound string        `note:"忽略无记录错误,yes/no"`
	Debug             *bool         `note:"是否开启debug日志"`
	Extension         []string      `note:"扩展配置"`
}
type Option func(*Options)

func WithHost(host string) Option {
	return func(opts *Options) {
		opts.Host = host
	}
}
func WithPort(port int) Option {
	return func(opts *Options) {
		opts.Port = port
	}
}
func WithUser(user string) Option {
	return func(opts *Options) {
		opts.User = user
	}
}
func WithPass(pass string) Option {
	return func(opts *Options) {
		opts.Pass = pass
	}
}
func WithDB(db string) Option {
	return func(opts *Options) {
		opts.DB = db
	}
}
func WithConfig(config string) Option {
	return func(opts *Options) {
		opts.Config = config
	}
}
func WithMaxIdle(maxIdle int) Option {
	return func(opts *Options) {
		opts.MaxIdle = maxIdle
	}
}
func WithMaxAlive(maxAlive int) Option {
	return func(opts *Options) {
		opts.MaxAlive = maxAlive
	}
}
func WithMaxAliveLife(maxAliveLife time.Duration) Option {
	return func(opts *Options) {
		opts.MaxAliveLife = maxAliveLife
	}
}
func WithLogSlow(logSlow time.Duration) Option {
	return func(opts *Options) {
		opts.LogSlow = logSlow
	}
}
func WithLogIgnoreNotFound(logIgnoreNotFound string) Option {
	return func(opts *Options) {
		opts.LogIgnoreNotFound = logIgnoreNotFound
	}
}
func WithDebug(debug bool) Option {
	return func(opts *Options) {
		opts.Debug = &debug
	}
}
func WithExtension(extension []string) Option {
	return func(opts *Options) {
		opts.Extension = extension
	}
}

func (o *Options) Validate() {
	if o.Debug == nil {
		o.Debug = zutil.Ptr(zutil.IsDebug())
	}
	if o.Host == "" {
		zlog.Fatalf("host is empty")
	}
	if o.Port == 0 {
		zlog.Fatalf("port is empty")
	}
	if o.User == "" {
		zlog.Fatalf("user is empty")
	}
	if o.Pass == "" {
		zlog.Fatalf("pass is empty")
	}
	if o.DB == "" {
		zlog.Fatalf("db is empty")
	}
}
func (o *Options) Dsn(database string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s %s",
		o.Host,
		o.Port,
		o.User,
		o.Pass,
		database,
		o.Config,
	)
}
