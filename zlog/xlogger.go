package zlog

import (
	"fmt"
)

type XLogger struct {
	l *Logger
}

func NewXLogger(options *Options) *XLogger {
	return &XLogger{
		l: NewZLogger(options),
	}
}

func (x *XLogger) Debug(args ...any) {
	x.l.Debugf(x.arrStr(args))
}
func (x *XLogger) Info(args ...any) {
	x.l.Infof(x.arrStr(args))
}
func (x *XLogger) Warn(args ...any) {
	x.l.Warnf(x.arrStr(args))
}
func (x *XLogger) Error(args ...any) {
	x.l.Errorf(x.arrStr(args))
}
func (x *XLogger) Debugf(format string, args ...any) {
	x.l.Debugf(format, args...)
}
func (x *XLogger) Infof(format string, args ...any) {
	x.l.Infof(format, args...)
}
func (x *XLogger) Warnf(format string, args ...any) {
	x.l.Warnf(format, args...)
}
func (x *XLogger) Errorf(format string, args ...any) {
	x.l.Errorf(format, args...)
}

func (x *XLogger) arrStr(args []any) string {
	str := ""
	for _, v := range args {
		str += fmt.Sprintf("%v ", v)
	}
	return str
}
