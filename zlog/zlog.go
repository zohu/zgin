package zlog

import "log/slog"

var zlog = NewZLogger(
	WithLevel(slog.LevelDebug),
	WithSkipCallers(1),
)

func WithOptions(opts ...Option) {
	zlog = NewZLogger(opts...)
}

func Debugf(format string, args ...any) {
	zlog.Debugf(format, args...)
}
func Infof(format string, args ...any) {
	zlog.Infof(format, args...)
}
func Warnf(format string, args ...any) {
	zlog.Warnf(format, args...)
}
func Errorf(format string, args ...any) {
	zlog.Errorf(format, args...)
}
func Fatalf(format string, args ...any) {
	zlog.Fatalf(format, args...)
}
func Panicf(format string, args ...any) {
	zlog.Panicf(format, args...)
}
