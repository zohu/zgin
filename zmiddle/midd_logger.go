package zmiddle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zauth"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
	"strings"
	"time"
)

type LoggerOptions struct {
	MaxBody int64    `yaml:"max_body"`
	MaxData int64    `yaml:"max_data"`
	Ignore  []string `yaml:"ignore"`
}

func (o *LoggerOptions) Validate() {
	o.MaxBody = zutil.FirstTruth(o.MaxBody, 1024)
	o.MaxData = zutil.FirstTruth(o.MaxData, 1024)
}

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
		zlog.Infof(buf.String())
	} else {
		zlog.Warnf(buf.String())
	}
}

func NewLogger(options *LoggerOptions) gin.HandlerFunc {
	options.Validate()
	return func(c *gin.Context) {
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

		if u, err := zauth.Auth(c); err != nil {
			item.Userid, item.Username = u.Userid(), u.UserName()
		}

		buf := zbuff.New()
		defer buf.Free()
		blw := &bodyWriter{body: buf, ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		item.Status = c.Writer.Status()
		item.Latency = time.Since(start).Milliseconds()

		item.Print()
	}
}

func formatBody(c *gin.Context) []byte {
	return nil
}
