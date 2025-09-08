package zdb

import (
	"context"
	"fmt"

	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zmap"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
)

var (
	p = zmap.New[*gorm.DB]()
	o *Options
)

func New(options *Options) {
	o = zutil.FirstTruth(options, new(Options))
	if err := o.Validate(); err != nil {
		zlog.Fatalf("validate options error: %v", err)
		return
	}

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

func AutoMigrate(dst []any) {
	if len(dst) > 0 {
		if err := NewDB(context.Background()).AutoMigrate(dst...); err != nil {
			zlog.Fatalf("auto migrate tables error: %v", err)
			return
		}
		zlog.Infof("auto migrate tables success")
	}
}

func NewDB(ctx context.Context, databases ...string) *gorm.DB {
	if len(databases) == 0 {
		databases = []string{o.DB}
	}
	if conn, ok := p.Get(databases[0]); ok {
		return conn
	}
	conn, err := newdb(o, databases[0])
	if err != nil {
		zlog.Fatalf("newdb error: %v", err)
		return nil
	}
	p.Set(databases[0], conn)
	return conn.WithContext(ctx)
}
