package zmiddle

import (
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zid"
	"github.com/zohu/zgin/zutil"
)

const RequestIdHeader = "X-Request-Id"

func NewRequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := zutil.FirstTruth(RequestId(c), zid.NextIdHex())
		c.Request.Header.Add(RequestIdHeader, rid)
		c.Header(RequestIdHeader, rid)
		c.Next()
	}
}
func RequestId(c *gin.Context) string {
	return zutil.FirstTruth(
		c.GetHeader(RequestIdHeader),
		c.Request.Header.Get(RequestIdHeader),
		c.Writer.Header().Get(RequestIdHeader),
	)
}
