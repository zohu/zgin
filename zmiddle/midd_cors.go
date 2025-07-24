package zmiddle

import (
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
	"slices"
)

type CorsOptions struct {
	AllowedOrigin    []string `yaml:"allowed_origin"`
	AllowedMethod    string   `yaml:"allowed_method"`
	AllowedHeader    string   `yaml:"allowed_header"`
	ExposeHeader     string   `yaml:"expose_header"`
	AllowCredentials string   `yaml:"allow_credentials"` // 如果存在，则允许发送Cookie，则Origin不可为*
}

func (o *CorsOptions) Validate() {
	if o.AllowCredentials == "true" && slices.Contains(o.AllowedOrigin, "*") {
		zlog.Fatal("allow_credentials is true, allowed_origin cannot be *")
	}
}
func NewCors(options *CorsOptions) gin.HandlerFunc {
	zlog.Infof("middleware cors enabled")
	options = zutil.FirstTruth(options, &CorsOptions{})
	options.Validate()
	return func(c *gin.Context) {
		if c.Request.Method == "" {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		origin := c.GetHeader("Origin")
		if origin != "" {
			if slices.Contains(options.AllowedOrigin, c.Request.URL.Host) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			if options.AllowedMethod != "" {
				c.Header("Access-Control-Allow-Methods", options.AllowedMethod)
			}
			if options.AllowedHeader != "" {
				c.Header("Access-Control-Allow-Headers", options.AllowedHeader)
			}
			if options.ExposeHeader != "" {
				c.Header("Access-Control-Expose-Headers", options.ExposeHeader)
			}
			if options.AllowCredentials != "" {
				c.Header("Access-Control-Allow-Credentials", options.AllowCredentials)
			}
			c.Header("Access-Control-Max-Age", "86400")
		}
		c.Next()
	}
}
