package zdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zohu/zgin/zlog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger struct {
	*zlog.Logger
	ignoreRecordNotFound bool
	logSlow              time.Duration
}

func NewLogger(o *Options) *Logger {
	options := &zlog.Options{SkipCallers: -1}
	if o.Debug != nil && *o.Debug {
		options.Level = slog.LevelDebug
	}
	l := &Logger{
		Logger:               zlog.NewZLogger(options),
		ignoreRecordNotFound: o.LogIgnoreNotFound == "yes",
		logSlow:              o.LogSlow,
	}
	logger.Default = l
	return l
}

func (l Logger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}
func (l Logger) Info(ctx context.Context, s string, i ...interface{}) {
	l.Infof(s, i...)
}
func (l Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.Warnf(s, i...)
}
func (l Logger) Error(ctx context.Context, s string, i ...interface{}) {
	l.Errorf(s, i...)
}
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!l.ignoreRecordNotFound || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.Errorf("rows=%d elapsed=%.3fs err=%s sql=%s", rows, elapsed.Seconds(), err.Error(), sql)
	case l.logSlow != 0 && elapsed > l.logSlow:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		l.Warnf("rows=%d elapsed=%.3fs %ssql=%s", rows, elapsed.Seconds(), e, sql)
	default:
		sql, rows := fc()
		var e string
		if err != nil {
			e = fmt.Sprintf("err=%s ", err.Error())
		}
		l.Debugf("rows=%d elapsed=%.3fs %ssql=%s", rows, elapsed.Seconds(), e, sql)
	}
}
