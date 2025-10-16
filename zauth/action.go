package zauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zlog"
)

func Action(actions ...string) gin.HandlerFunc {
	for _, action := range actions {
		if len(strings.Split(action, ":")) != 3 {
			zlog.Fatalf("接口权限定义错误，应为[*:*:*]格式")
		}
	}
	return func(c *gin.Context) {
		if auth, ok := Auth(c); ok {
			if require(c.Request.Context(), auth.Userid(), actions) {
				c.Next()
				return
			}
		}
		zgin.AbortHttpCode(c, http.StatusForbidden, zgin.MessageActionInvalid.Resp(c))
	}
}

func require(ctx context.Context, userid string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	per := LoadPermission(ctx, userid)
	for _, pattern := range patterns {
		if !per.Match(pattern) {
			return false
		}
	}
	return true
}
