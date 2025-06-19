package zmiddle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zauth"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"io"
	"net/http"
	"strings"
	"time"
)

type LoggerOptions struct {
	MaxBody int      `yaml:"max_body"`
	MaxData int      `yaml:"max_data"`
	Ignore  []string `yaml:"ignore"`
}

func (o *LoggerOptions) Validate() {
	o.MaxBody = zutil.FirstTruth(o.MaxBody, 1024)
	o.MaxData = zutil.FirstTruth(o.MaxData, 1024)
}

var mLogger = zlog.NewZLogger(&zlog.Options{
	SkipCallers: -1,
})

/**
 * format: method status ip latency request_id path userid-username query body >>> data error
 */

type LoggerItem struct {
	Method    string `json:"method"`
	Status    int    `json:"status"`
	Ip        string `json:"ip"`
	Latency   int64  `json:"latency"`
	RequestId string `json:"request_id"`
	Path      string `json:"path"`
	Userid    string `json:"userid"`
	Username  string `json:"username"`
	Query     string `json:"query"`
	Body      string `json:"body"`
	Data      string `json:"data"`
}

func (l *LoggerItem) Print() {
	buf := zbuff.New()
	defer buf.Free()

	_, _ = buf.WriteStringIf(l.Method != "", fmt.Sprintf("%s ", l.Method))
	_, _ = buf.WriteStringIf(l.Status != 0, fmt.Sprintf("%d ", l.Status))
	_, _ = buf.WriteStringIf(l.Ip != "", fmt.Sprintf("%s ", l.Ip))
	buf.WriteString(fmt.Sprintf("%dms ", l.Latency))
	_, _ = buf.WriteStringIf(l.RequestId != "", fmt.Sprintf("%s ", l.RequestId))
	_, _ = buf.WriteStringIf(l.Path != "", fmt.Sprintf("%s ", l.Path))
	_, _ = buf.WriteStringIf(l.Userid != "", fmt.Sprintf("%s-%s ", l.Userid, l.Username))
	_, _ = buf.WriteStringIf(l.Query != "", fmt.Sprintf("%s ", l.Query))
	_, _ = buf.WriteStringIf(l.Body != "", fmt.Sprintf("%s ", l.Body))
	_, _ = buf.WriteStringIf(l.Data != "", fmt.Sprintf(">>> %s", l.Data))

	if l.Status == http.StatusOK {
		mLogger.Infof(buf.String())
	} else {
		mLogger.Warnf(buf.String())
	}
}

func NewLogger(options *LoggerOptions) gin.HandlerFunc {
	zlog.Infof("middleware api logger enabled")
	options = zutil.FirstTruth(options, &LoggerOptions{})
	options.Validate()
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		for _, ignore := range options.Ignore {
			if strings.HasPrefix(c.Request.URL.Path[1:], ignore) {
				c.Next()
				return
			}
		}
		start := time.Now()
		item := &LoggerItem{
			Method:    c.Request.Method,
			Ip:        c.ClientIP(),
			RequestId: RequestId(c),
			Path:      c.Request.URL.Path,
			Query:     c.Request.URL.Query().Encode(),
		}

		if u, ok := zauth.Auth(c); ok {
			item.Userid, item.Username = u.Userid(), u.UserName()
		}

		buf := zbuff.New()
		defer buf.Free()
		blw := &bodyWriter{body: buf, ResponseWriter: c.Writer}
		c.Writer = blw

		// body
		{
			if c.ContentType() == gin.MIMEMultipartPOSTForm {
				item.Body = "upload file..."
			} else {
				body, _ := c.GetRawData()
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				if len(body) > 0 {
					if body[0] == 123 && body[len(body)-1] == 125 {
						dst := zbuff.New()
						defer dst.Free()
						_ = json.Compact(dst.Buffer, body)
						body = dst.Bytes()
					}
					if len(body) > options.MaxBody {
						body = body[:options.MaxBody]
					}
					item.Body = string(body)
				}
			}
		}

		c.Next()

		// data
		{
			data := blw.body.Bytes()
			if len(data) > 0 {
				if data[0] == 123 && data[len(data)-1] == 125 {
					dst := zbuff.New()
					defer dst.Free()
					_ = json.Compact(dst.Buffer, data)
					data = dst.Bytes()
				}
				if len(data) > options.MaxData {
					data = data[:options.MaxData]
				}
				item.Data = string(data)
			}
		}

		item.Status = c.Writer.Status()
		item.Latency = time.Since(start).Milliseconds()

		item.Print()
	}
}
