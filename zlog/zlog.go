package zlog

import "log/slog"

var zlog = NewZLogger(&Options{
	Level:       slog.LevelDebug,
	SkipCallers: 1,
})

func WithOptions(options *Options) {
	zlog = NewZLogger(options)
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
