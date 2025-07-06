package zauth

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zch"
	"github.com/zohu/zgin/zlog"
	"net/http"
	"strings"
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
			}
		}
		zgin.AbortHttpCode(c, http.StatusForbidden, zgin.MessageActionInvalid.Resp(c))
	}
}

func require(ctx context.Context, userid string, actions []string) bool {
	if len(actions) == 0 {
		return true
	}
	vKey := zch.PrefixAuthAction.Key(userid)
	if zch.R().Exists(ctx, vKey).Val() > 0 {
		zch.R().Expire(ctx, vKey, options.Age)
		if zch.R().SIsMember(ctx, vKey, "*:*:*").Val() {
			return true
		}
		for _, action := range actions {
			item := strings.Split(action, ":")
			allow := []string{
				action,
				fmt.Sprintf("%s:*:*", item[0]),
				fmt.Sprintf("*:%s:*", item[1]),
				fmt.Sprintf("*:*:%s", item[2]),
				fmt.Sprintf("%s:%s:*", item[0], item[1]),
				fmt.Sprintf("%s:*:%s", item[0], item[2]),
				fmt.Sprintf("*:%s:%s", item[1], item[2]),
			}
			for _, s := range allow {
				if zch.R().SIsMember(ctx, vKey, s).Val() {
					return true
				}
			}
		}
	}
	return false
}
