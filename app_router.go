package zgin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NoRoute(c *gin.Context) {
	AbortHttpCode(c, http.StatusNotFound, MessageInvalidPath.Resp(c))
}
func NoMethod(c *gin.Context) {
	AbortHttpCode(c, http.StatusMethodNotAllowed, MessageInvalidMethod.Resp(c))
}
func Health(c *gin.Context) {
	AbortString(c, "ok")
}
func NotImplemented(c *gin.Context) {
	AbortHttpCode(c, http.StatusNotImplemented, MessageNotImplemented.Resp(c))
}
