package zmiddle

import (
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zlog"
	"net/http"
)

func NewRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zlog.Errorf("server panic: %v", err)
				zgin.AbortHttpCode(c, http.StatusInternalServerError, zgin.MessageInvalidRequest.Resp(c))
			}
		}()
		c.Next()
	}
}
