package zmiddle

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zbuff"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
)

type TimeoutOptions struct {
	Timeout time.Duration `yaml:"timeout"`
	Exclude []string      `yaml:"exclude" note:"不超时的PATH"`
}

func (o *TimeoutOptions) Validate() {
	o.Timeout = zutil.FirstTruth(o.Timeout, time.Minute)
}

type bodyWriter struct {
	gin.ResponseWriter
	body *zbuff.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewTimeout(options *TimeoutOptions) gin.HandlerFunc {
	zlog.Infof("middleware timeout enabled")
	options = zutil.FirstTruth(options, &TimeoutOptions{})
	options.Validate()

	return func(c *gin.Context) {
		for _, exclude := range options.Exclude {
			if strings.HasPrefix(strings.TrimPrefix(c.Request.URL.Path, "/"), strings.TrimPrefix(exclude, "/")) {
				c.Next()
				return
			}
		}
		buf := zbuff.New()
		defer buf.Free()
		blw := &bodyWriter{body: buf, ResponseWriter: c.Writer}
		c.Writer = blw

		ctx, cancel := context.WithTimeout(c.Request.Context(), options.Timeout)
		c.Request = c.Request.WithContext(ctx)
		defer cancel()

		fChan := make(chan struct{}, 1)    // finish chan
		pChan := make(chan interface{}, 1) // panic chan
		go func() {
			defer func() {
				if err := recover(); err != nil {
					pChan <- fmt.Sprintf("%v\n%s", err, debug.Stack())
				}
			}()
			c.Next()
			fChan <- struct{}{}
		}()

		select {
		case <-ctx.Done():
			zgin.AbortHttpCode(c, http.StatusGatewayTimeout, zgin.MessageTimeout.Resp(c))
		case <-fChan:
			return
		case err := <-pChan:
			zlog.Errorf("panic: %v", err)
			zgin.AbortHttpCode(c, http.StatusInternalServerError, zgin.MessageRequestInvalid.Resp(c))
		}
	}
}
