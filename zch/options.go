package zch

import (
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"time"
)

type Options struct {
	Expiration    time.Duration
	CleanInterval time.Duration
	Addrs         []string
	Database      int
	Password      string
	Prefix        string
	ClientName    string
}

type Option func(*Options)

func WithExpiration(expiration time.Duration) Option {
	return func(opts *Options) {
		opts.Expiration = expiration
	}
}
func WithCleanInterval(cleanInterval time.Duration) Option {
	return func(opts *Options) {
		opts.CleanInterval = cleanInterval
	}
}
func WithAddrs(addrs []string) Option {
	return func(opts *Options) {
		opts.Addrs = addrs
	}
}
func WithDatabase(database int) Option {
	return func(opts *Options) {
		opts.Database = database
	}
}
func WithPassword(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}
func WithPrefix(prefix string) Option {
	return func(opts *Options) {
		opts.Prefix = prefix
	}
}
func WithClientName(clientName string) Option {
	return func(opts *Options) {
		opts.ClientName = clientName
	}
}

func (o *Options) Validate() {
	if len(o.Addrs) == 0 {
		zlog.Fatalf("addrs is empty")
	}
	if o.Expiration == 0 {
		o.Expiration = time.Hour
	}
	if o.CleanInterval == 0 {
		o.CleanInterval = time.Minute * 5
	}
	o.ClientName = zutil.FirstTruth(o.ClientName, "zch")
}
