package zauth

import (
	"encoding/base64"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zch"
	"github.com/zohu/zgin/zcpt"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
	"strings"
)

const (
	AESKey              = "315c2wd6vpc7q4hx"
	LocalsUserPrefix    = "user"
	LocalsSessionPrefix = "session"
)

var options *Options

func NewMiddleware[T Userinfo](opts *Options) gin.HandlerFunc {
	opts = zutil.FirstTruth(opts, &Options{})
	if err := opts.Validate(); err != nil {
		zlog.Fatalf("options is invalid: %v", err)
	}
	options = opts
	return func(c *gin.Context) {
		// 路径校验
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		if options.PathSkip(c.Request.URL.Path) {
			c.Next()
			return
		}
		for _, ignore := range options.WhiteList {
			if strings.HasPrefix(c.Request.URL.Path[1:], ignore) {
				c.Next()
				return
			}
		}
		// 校验登录态
		token := Token(c)
		if token == "" {
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		// 解析登录态
		d, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			zlog.Warnf("auth token decode err: %v", err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		d, err = zcpt.AesDecryptCBC(d, []byte(AESKey))
		if err != nil {
			zlog.Warnf("auth token decrypt err: %v", err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		tks := strings.Split(string(d), "##")
		if len(tks) != 5 {
			zlog.Warnf("auth token len err: 5 != [%d]", len(tks))
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}

		agent := tks[1]
		ip := tks[2]
		userid := tks[3]
		// 校验UA是否变化
		if !options.AllowUaChange && agent != zcpt.Md5(c.Request.UserAgent()) {
			zlog.Warnf("auth token userid=%s ua changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		// 校验IP是否变化
		if !options.AllowIpChange && ip != c.ClientIP() {
			zlog.Warnf("auth token userid=%s ip changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		// 提取用户数据
		vKey := zch.PrefixAuthToken.Key(userid)
		uStr := zch.R().Get(c.Request.Context(), vKey).Val()
		if uStr == "" {
			zlog.Warnf("auth token userid=%s not found", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		var auth Authorization[T]
		if err = sonic.UnmarshalString(uStr, &auth); err != nil {
			zlog.Warnf("auth token userid=%s unmarshal err: %v", userid, err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginTokenInvalid.Resp(c))
			return
		}
		// 是否允许多设备登录
		if !options.AllowMultipleDevice && auth.Session != zcpt.Md5(token) {
			zlog.Warnf("auth token userid=%s device changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.MessageLoginSessionInvalid.Resp(c))
			return
		}
		// 用户状态是否正常
		if vali := auth.Value.Validate(); vali != zgin.MessageSuccess {
			zlog.Warnf("auth token userid=%s status invalid: %s", userid, vali)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, vali.Resp(c))
			return
		}

		// 临时存储用户资料
		c.Set(LocalsUserPrefix, zutil.Ptr(auth.Value))
		c.Set(LocalsSessionPrefix, auth.Session)

		// 刷新Token有效期
		c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
		zch.R().Set(c.Request.Context(), vKey, uStr, options.Age)
		c.Next()
	}
}

func UpdateAuth(c *gin.Context, user Userinfo) {
	session, ok := c.Get(LocalsSessionPrefix)
	if !ok {
		token := Token(c)
		if token == "" {
			return
		}
		session = zcpt.Md5(token)
	}
	c.Set(LocalsUserPrefix, zutil.Ptr(user))
	uStr, _ := sonic.MarshalString(&Authorization[Userinfo]{Session: session.(string), Value: user})
	vKey := zch.PrefixAuthToken.Key(user.Userid())
	zch.R().Set(c.Request.Context(), vKey, uStr, options.Age)
}

func Auth(c *gin.Context) (Userinfo, bool) {
	if u, ok := c.Get(LocalsUserPrefix); ok {
		return u.(Userinfo), ok
	}
	return nil, false
}

func Token(c *gin.Context) string {
	auth, _ := c.Cookie("auth")
	return strings.TrimSpace(zutil.FirstTruth(
		c.GetHeader("Authorization"),
		c.Query("auth"),
		auth,
	))
}
