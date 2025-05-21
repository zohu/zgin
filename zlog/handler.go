package zlog

import (
	"context"
	"encoding"
	"fmt"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zutil"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode"
)

/**
 * slog.Handler implementation
 */

const (
	ANSIReset  = "\033[0m"
	ANSIDebug  = "\033[90m"
	ANSIInfo   = "\033[32m"
	ANSIWarn   = "\033[33m"
	ANSIError  = "\033[31m"
	ANSITime   = "\033[37m"
	ANSISource = "\033[34;4m"
)

type Options struct {
	SkipCallers int          `yaml:"skip_callers"`
	Level       slog.Leveler `yaml:"level"`
	TimeFormat  string       `yaml:"time_format"`
	NoColor     bool         `yaml:"no_color"`
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr
	Writer      io.Writer
}

func (o *Options) Validate() {
	o.Level = zutil.FirstTruth(o.Level, slog.LevelInfo)
	o.TimeFormat = zutil.FirstTruth(o.TimeFormat, time.DateTime)
	if o.ReplaceAttr == nil {
		o.ReplaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			return attr
		}
	}
	if o.Writer == nil {
		o.Writer = os.Stdout
	}
}

func NewHandler(options *Options) slog.Handler {
	options = zutil.FirstTruth(options, &Options{})
	options.Validate()

	h := new(handler)
	h.level = options.Level
	h.timeFormat = options.TimeFormat
	h.skipCallers = options.SkipCallers
	h.replaceAttr = options.ReplaceAttr
	h.timeFormat = options.TimeFormat
	h.noColor = options.NoColor
	h.writer = options.Writer
	return h
}

type handler struct {
	attrsPrefix string
	groupPrefix string
	groups      []string

	skipCallers int
	level       slog.Leveler
	replaceAttr func(groups []string, attr slog.Attr) slog.Attr
	timeFormat  string
	noColor     bool

	writer io.Writer
	sync.Mutex
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}
func (h *handler) Handle(_ context.Context, r slog.Record) error {
	buf := zbuff.New()
	defer buf.Free()

	rep := h.replaceAttr

	// time
	if !r.Time.IsZero() {
		h.color(buf, ANSITime)
		val := r.Time.Round(0) // strip monotonic to match Attr behavior
		if rep == nil {
			buf.WriteString(r.Time.Format(h.timeFormat))
			_ = buf.WriteByte(' ')
		} else if a := rep(nil /* groups */, slog.Time(slog.TimeKey, val)); a.Key != "" {
			if a.Value.Kind() == slog.KindTime {
				buf.WriteString(a.Value.Time().Format(h.timeFormat))
			} else {
				h.appendValue(buf, a.Value, false)
			}
			_ = buf.WriteByte(' ')
		}
		h.colorEnd(buf)
	}

	// level
	h.colorLevel(buf, r.Level)
	if rep == nil {
		h.appendLevel(buf, r.Level)
		_ = buf.WriteByte(' ')
	} else if a := rep(nil /* groups */, slog.Any(slog.LevelKey, r.Level)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
		_ = buf.WriteByte(' ')
	}
	h.colorEnd(buf)

	// source
	if h.skipCallers >= 0 {
		pcs := make([]uintptr, 16)
		n := runtime.Callers(6+h.skipCallers, pcs)
		fs := runtime.CallersFrames(pcs[:n])
		f, _ := fs.Next()
		if f.File != "" {
			src := &slog.Source{
				Function: f.Function,
				File:     f.File,
				Line:     f.Line,
			}
			h.color(buf, ANSISource)
			if rep == nil {
				h.appendSource(buf, src)
			} else if a := rep(nil /* groups */, slog.Any(slog.SourceKey, src)); a.Key != "" {
				h.appendValue(buf, a.Value, false)
			}
			h.colorEnd(buf)
			_ = buf.WriteByte(' ')
		}
	}

	// message
	h.colorLevel(buf, r.Level)
	if rep == nil {
		_, _ = buf.WriteString(r.Message)
	} else if a := rep(nil /* groups */, slog.String(slog.MessageKey, r.Message)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
	}
	h.colorEnd(buf)
	_ = buf.WriteByte(' ')

	// handler attributes
	if len(h.attrsPrefix) > 0 {
		_, _ = buf.WriteString(h.attrsPrefix)
	}

	// attributes
	r.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
		return true
	})

	if buf.Len() == 0 {
		return nil
	}

	// replace last space with newline
	buf.Truncate(buf.Len() - 1)
	buf.WriteByte('\n')

	h.Lock()
	defer h.Unlock()

	_, err := h.writer.Write(buf.Bytes())
	return err
}
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := h.clone()

	buf := zbuff.New()
	defer buf.Free()

	// write attributes to buffer
	for _, attr := range attrs {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
	}
	h2.attrsPrefix = h.attrsPrefix + buf.String()
	return h2
}
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groupPrefix += name + "."
	h2.groups = append(h2.groups, name)
	return h2
}
func appendLevelDelta(buf *zbuff.Buffer, delta slog.Level) {
	if delta == 0 {
		return
	} else if delta > 0 {
		_ = buf.WriteByte('+')
	}
	buf.WriteString(strconv.Itoa(int(delta)))
}
func (h *handler) appendKey(buf *zbuff.Buffer, key, groups string) {
	h.appendString(buf, groups+key, true)
	_ = buf.WriteByte('=')
}
func (h *handler) appendLevel(buf *zbuff.Buffer, level slog.Level) {
	switch {
	case level < slog.LevelInfo:
		_, _ = buf.WriteString("DBG")
		appendLevelDelta(buf, level-slog.LevelDebug)
	case level < slog.LevelWarn:
		_, _ = buf.WriteString("INF")
		appendLevelDelta(buf, level-slog.LevelInfo)
	case level < slog.LevelError:
		_, _ = buf.WriteString("WRN")
		appendLevelDelta(buf, level-slog.LevelWarn)
	default:
		_, _ = buf.WriteString("ERR")
		appendLevelDelta(buf, level-slog.LevelError)
	}
}
func (h *handler) appendValue(buf *zbuff.Buffer, v slog.Value, quote bool) {
	switch v.Kind() {
	case slog.KindString:
		h.appendString(buf, v.String(), quote)
	case slog.KindInt64:
		buf.WriteString(strconv.FormatInt(v.Int64(), 10))
	case slog.KindUint64:
		buf.WriteString(strconv.FormatUint(v.Uint64(), 10))
	case slog.KindFloat64:
		buf.WriteString(strconv.FormatFloat(v.Float64(), 'g', -1, 64))
	case slog.KindBool:
		buf.WriteString(strconv.FormatBool(v.Bool()))
	case slog.KindDuration:
		h.appendString(buf, v.Duration().String(), quote)
	case slog.KindTime:
		h.appendString(buf, v.Time().String(), quote)
	case slog.KindAny:
		switch cv := v.Any().(type) {
		case slog.Level:
			h.appendLevel(buf, cv)
		case encoding.TextMarshaler:
			data, err := cv.MarshalText()
			if err != nil {
				break
			}
			h.appendString(buf, string(data), quote)
		case *slog.Source:
			h.appendSource(buf, cv)
		default:
			h.appendString(buf, fmt.Sprintf("%+v", v.Any()), quote)
		}
	default:
	}
}
func (h *handler) appendAttr(buf *zbuff.Buffer, attr slog.Attr, groupsPrefix string, groups []string) {
	attr.Value = attr.Value.Resolve()
	if rep := h.replaceAttr; rep != nil && attr.Value.Kind() != slog.KindGroup {
		attr = rep(groups, attr)
		attr.Value = attr.Value.Resolve()
	}

	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			groupsPrefix += attr.Key + "."
			groups = append(groups, attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(buf, groupAttr, groupsPrefix, groups)
		}
		return
	}

	h.appendKey(buf, attr.Key, groupsPrefix)
	h.appendValue(buf, attr.Value, true)
	_ = buf.WriteByte(' ')
}

func (h *handler) color(buf *zbuff.Buffer, ansi string) {
	if h.noColor {
		return
	}
	_, _ = buf.WriteString(ansi)
}
func (h *handler) colorLevel(buf *zbuff.Buffer, level slog.Level) {
	if h.noColor {
		return
	}
	switch level {
	case slog.LevelDebug:
		_, _ = buf.WriteString(ANSIDebug)
	case slog.LevelInfo:
		_, _ = buf.WriteString(ANSIInfo)
	case slog.LevelWarn:
		_, _ = buf.WriteString(ANSIWarn)
	case slog.LevelError:
		_, _ = buf.WriteString(ANSIError)
	default:
		_, _ = buf.WriteString(ANSIDebug)
	}
}
func (h *handler) colorEnd(buf *zbuff.Buffer) {
	if h.noColor {
		return
	}
	_, _ = buf.WriteString(ANSIReset)
}

func (h *handler) appendSource(buf *zbuff.Buffer, src *slog.Source) {
	dir, file := filepath.Split(src.File)
	_, _ = buf.WriteString(filepath.Join(filepath.Base(dir), file))
	_ = buf.WriteByte(':')
	_, _ = buf.WriteString(strconv.Itoa(src.Line))
}
func (h *handler) appendString(buf *zbuff.Buffer, s string, quote bool) {
	if quote && h.needsQuote(s) {
		buf.WriteByte('"')
		buf.WriteString(s)
		buf.WriteByte('"')
	} else {
		_, _ = buf.WriteString(s)
	}
}
func (h *handler) needsQuote(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}
func (h *handler) clone() *handler {
	return &handler{
		attrsPrefix: h.attrsPrefix,
		groupPrefix: h.groupPrefix,
		groups:      h.groups,
		writer:      h.writer,
		skipCallers: h.skipCallers,
		level:       h.level,
		replaceAttr: h.replaceAttr,
		timeFormat:  h.timeFormat,
		noColor:     h.noColor,
	}
}
