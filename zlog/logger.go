package zlog

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

/**
 * safe writer for log
 */

const maxTokenLength = bufio.MaxScanTokenSize / 2

func scanLinesOrGiveLong(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	if advance > 0 || token != nil || err != nil {
		return
	}
	if len(data) < maxTokenLength {
		return
	}
	return maxTokenLength, data[0:maxTokenLength], nil
}

func writerFinalizer(w *io.PipeWriter) {
	_ = w.Close()
}

func scan(r *io.PipeReader, fn func(string, ...any)) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLinesOrGiveLong)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			fn(text)
		}
	}
	_ = r.Close()
}
func SafeWriter(fn func(string, ...any)) *io.PipeWriter {
	reader, writer := io.Pipe()
	go scan(reader, fn)
	runtime.SetFinalizer(writer, writerFinalizer)
	return writer
}

/**
 * logger
 */

type Logger struct {
	s *slog.Logger
}

func NewZLogger(opts ...Option) *Logger {
	return &Logger{
		s: slog.New(NewHandler(opts...)),
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.s.Debug(msg, args...)
}
func (l *Logger) Print(msg string, args ...any) {
	l.s.Info(msg, args...)
}
func (l *Logger) Info(msg string, args ...any) {
	l.s.Info(msg, args...)
}
func (l *Logger) Warn(msg string, args ...any) {
	l.s.Warn(msg, args...)
}
func (l *Logger) Error(msg string, args ...any) {
	l.s.Error(msg, args...)
}
func (l *Logger) Fatal(msg string, args ...any) {
	l.s.Error(msg, args...)
	os.Exit(1)
}
func (l *Logger) Panic(msg string, args ...any) {
	l.s.Error(msg, args...)
	panic(msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}
func (l *Logger) Printf(format string, args ...any) {
	l.Print(fmt.Sprintf(format, args...))
}
func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}
func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}
func (l *Logger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}
func (l *Logger) Panicf(format string, args ...any) {
	l.Panic(fmt.Sprintf(format, args...))
}

func (l *Logger) SafeInfof(format string, args ...any) {
	w := SafeWriter(l.Infof)
	_, _ = w.Write([]byte(fmt.Sprintf(format, args...)))
}
