package zgin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewRespWithData(c *gin.Context, data any) *RespBean {
	return MessageSuccess.Resp(c).WithData(data)
}
func NewRespWithList[T any](c *gin.Context, data *RespListBean[T]) *RespBean {
	return NewRespWithData(c, data)
}

func AbortHttpCode(c *gin.Context, code int, resp *RespBean) {
	switch c.ContentType() {
	case gin.MIMEPlain:
		c.String(code, fmt.Sprintf("%s", resp.Data))
	case gin.MIMEXML, gin.MIMEXML2:
		c.XML(code, resp)
	default:
		c.JSON(code, resp)
	}
	c.Abort()
}
func AbortString(c *gin.Context, message string) {
	c.String(http.StatusOK, message)
	c.Abort()
}
func Abort(c *gin.Context, resp *RespBean) {
	AbortHttpCode(c, http.StatusOK, resp)
}
