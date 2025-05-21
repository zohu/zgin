package zmiddle

import (
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
)

type CorsOptions struct {
	AllowedOrigin    string `yaml:"allowed_origin"`
	AllowedMethod    string `yaml:"allowed_method"`
	AllowedHeader    string `yaml:"allowed_header"`
	ExposeHeader     string `yaml:"expose_header"`
	AllowCredentials string `yaml:"allow_credentials"`
}

func (o *CorsOptions) Validate() {

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
		if options.AllowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", options.AllowedOrigin)
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
		c.Next()
	}
}
