package zdb

import (
	"context"
	"fmt"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zmap"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
	"time"
)

var (
	p zmap.ConcurrentMap[string, *gorm.DB]
	o *Options
)

func New(opts ...Option) {
	o = &Options{
		Config:            "sslmode=disable TimeZone=Asia/Shanghai",
		MaxIdle:           10,
		MaxAlive:          100,
		MaxAliveLife:      time.Hour,
		LogSlow:           time.Second * 5,
		LogIgnoreNotFound: "yes",
		Debug:             zutil.Ptr(zutil.IsDebug()),
	}
	for _, opt := range opts {
		opt(o)
	}
	o.Validate()

	db := NewDB(context.Background())
	// init extension
	for _, ext := range o.Extension {
		if err := db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s;", ext)).Error; err != nil {
			zlog.Fatalf("register extension [%s] error: %v", ext, err)
			return
		}
		zlog.Infof("register extension [%s] success", ext)
	}
}

func AutoMigrate(dst ...any) {
	if err := NewDB(context.Background()).AutoMigrate(dst...); err != nil {
		zlog.Fatalf("auto migrate tables error: %v", err)
		return
	}
	zlog.Infof("auto migrate tables success")
}

func NewDB(ctx context.Context, databases ...string) *gorm.DB {
	database := zutil.When(len(databases) > 0, databases[0], o.DB)
	if conn, ok := p.Get(database); ok {
		return conn
	}
	conn, err := newdb(o, database)
	if err != nil {
		zlog.Fatalf("newdb error: %v", err)
		return nil
	}
	p.Set(database, conn)
	return conn.WithContext(ctx)
}
