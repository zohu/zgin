package zmiddle

import (
	"fmt"
	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zutil"
	"github.com/zohu/zlog"
	"io"
	"net/http"
	"time"
)

type LimitOptions struct {
	BodySize int     `yaml:"body_size" note:"MB"`
	Rate     float64 `yaml:"rate" note:"every minute"`
}

func (o *LimitOptions) Validate() {
	o.BodySize = zutil.FirstTruth(o.BodySize, 2)
	o.Rate = zutil.FirstTruth(o.Rate, 9999)
}

func NewLimit(options *LimitOptions) gin.HandlerFunc {
	zlog.Infof("middleware limit enabled")
	options = zutil.FirstTruth(options, &LimitOptions{})
	options.Validate()

	lmt := tollbooth.NewLimiter(options.Rate, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})
	lmt.SetIPLookup(limiter.IPLookup{
		Name:           "X-Real-IP",
		IndexFromRight: 0,
	})

	return func(c *gin.Context) {
		if httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request); httpError != nil {
			_ = c.AbortWithError(httpError.StatusCode, httpError)
			return
		}
		c.Request.Body = &limitBody{
			ctx:       c,
			r:         c.Request.Body,
			remaining: options.BodySize * 1024 * 1024,
		}
		c.Next()
	}
}

type limitBody struct {
	ctx       *gin.Context
	r         io.ReadCloser
	remaining int
	aborted   bool
	eofed     bool
}

func (l *limitBody) TooLarge() (int, error) {
	err := fmt.Errorf("HTTP request body too large")
	if !l.aborted {
		l.aborted = true
		l.ctx.Header("Connection", "close")
		_ = l.ctx.AbortWithError(http.StatusRequestEntityTooLarge, err)
	}
	return 0, err
}

func (l *limitBody) Read(p []byte) (n int, err error) {
	remaining := l.remaining
	if l.remaining == 0 {
		if l.eofed {
			return l.TooLarge()
		}
		remaining = 1
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	n, err = l.r.Read(p)
	if err == io.EOF {
		l.eofed = true
	}
	if l.remaining == 0 {
		if n > 0 {
			return l.TooLarge()
		}
		return n, err
	}
	l.remaining -= n
	if l.remaining < 0 {
		l.remaining = 0
	}
	return n, err
}
func (l *limitBody) Close() error {
	return l.r.Close()
}
