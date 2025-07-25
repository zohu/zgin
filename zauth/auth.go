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
	"net/url"
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

		var auth Authorization[T]
		if msgID := ScanAuth(c, &auth); msgID != zgin.MessageSuccess {
			zgin.AbortHttpCode(c, http.StatusUnauthorized, msgID.Resp(c))
			return
		}

		// 临时存储用户资料
		c.Set(LocalsUserPrefix, zutil.Ptr(auth.Value))
		c.Set(LocalsSessionPrefix, auth.Session)

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
	query, _ := url.QueryUnescape(c.Query("auth"))
	return strings.TrimSpace(zutil.FirstTruth(
		c.GetHeader("Authorization"),
		query,
		auth,
	))
}

func ScanAuth[T Userinfo](c *gin.Context, auth *Authorization[T]) zgin.MessageID {
	token := Token(c)
	if token == "" {
		return zgin.MessageLoginTokenInvalid
	}
	// 解析登录态
	d, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		zlog.Warnf("auth token decode err: %v, token=%s", err, token)
		return zgin.MessageLoginTokenInvalid
	}
	d, err = zcpt.AesDecryptCBC(d, []byte(AESKey))
	if err != nil {
		zlog.Warnf("auth token decrypt err: %v, token=%s", err, token)
		return zgin.MessageLoginTokenInvalid
	}
	tks := strings.Split(string(d), "##")
	if len(tks) != 5 {
		zlog.Warnf("auth token len err: 5 != [%d]", len(tks))
		return zgin.MessageLoginTokenInvalid
	}

	agent := tks[1]
	ip := tks[2]
	userid := tks[3]
	// 校验UA是否变化
	if !options.AllowUaChange && agent != zcpt.Md5(c.Request.UserAgent()) {
		zlog.Warnf("auth token userid=%s ua changed", userid)
		return zgin.MessageLoginTokenInvalid
	}
	// 校验IP是否变化
	if !options.AllowIpChange && ip != c.ClientIP() {
		zlog.Warnf("auth token userid=%s ip changed", userid)
		return zgin.MessageLoginTokenInvalid
	}
	// 提取用户数据
	vKey := zch.PrefixAuthToken.Key(userid)
	uStr := zch.R().Get(c.Request.Context(), vKey).Val()
	if uStr == "" {
		zlog.Warnf("auth token userid=%s not found", userid)
		return zgin.MessageLoginTokenInvalid
	}
	if err = sonic.UnmarshalString(uStr, &auth); err != nil {
		zlog.Warnf("auth token userid=%s unmarshal err: %v", userid, err)
		return zgin.MessageLoginTokenInvalid
	}
	// 是否允许多设备登录
	if !options.AllowMultipleDevice && auth.Session != zcpt.Md5(token) {
		zlog.Warnf("auth token userid=%s device changed", userid)
		return zgin.MessageLoginSessionInvalid
	}
	// 用户状态是否正常
	if vali := auth.Value.Validate(); vali != zgin.MessageSuccess {
		zlog.Warnf("auth token userid=%s status invalid: %s", userid, vali)
		return vali
	}
	// 刷新Token有效期
	c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
	zch.R().Set(c.Request.Context(), vKey, uStr, options.Age)
	return zgin.MessageSuccess
}
