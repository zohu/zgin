package zgin

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
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
		v, _ := sonic.MarshalString(resp)
		c.String(code, v)
	case gin.MIMEXML, gin.MIMEXML2:
		c.XML(code, resp)
	default:
		c.JSON(code, resp)
	}
	if resp.Code != 1 {
		c.Set("__CODE__", resp.Code)
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
