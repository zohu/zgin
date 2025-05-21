package zauth

import (
	"encoding/base64"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zcpt"
	"github.com/zohu/zgin/zid"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"net/http"
	"strings"
	"time"
)

type Userinfo interface {
	Userid() string
	UserName() string
	Validate() error
}

type Authorization struct {
	Session string   `json:"session"`
	Value   Userinfo `json:"value"`
}

const (
	AESKey              = "315c2wd6vpc7q4hx"
	TokenPrefix         = "auth"
	LocalsUserPrefix    = "user"
	LocalsSessionPrefix = "session"
)

var options *Options

func New(opts *Options) gin.HandlerFunc {
	opts = zutil.FirstTruth(opts, &Options{})
	if err := opts.Validate(); err != nil {
		zlog.Fatalf("options is invalid: %v", err)
	}
	options = opts
	return func(c *gin.Context) {
		// 路径校验
		if options.PathSkip(strings.TrimPrefix(c.Request.URL.Path, "/")) {
			c.Next()
			return
		}
		// 校验登录态
		token := Token(c)
		if token == "" {
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		// 解析登录态
		d, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			zlog.Warnf("auth token decode err: %v", err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		d, err = zcpt.AesDecryptCBC(d, []byte(AESKey))
		if err != nil {
			zlog.Warnf("auth token decrypt err: %v", err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		tks := strings.Split(string(d), "##")
		if len(tks) != 5 {
			zlog.Warnf("auth token len err: 5 != [%d]", len(tks))
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}

		agent := tks[1]
		ip := tks[2]
		userid := tks[3]
		// 校验UA是否变化
		if !options.AllowUaChange && agent != zcpt.Md5(c.Request.UserAgent()) {
			zlog.Warnf("auth token userid=%s ua changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		// 校验IP是否变化
		if !options.AllowIpChange && ip != c.ClientIP() {
			zlog.Warnf("auth token userid=%s ip changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		// 提取用户数据
		vKey := options.WithPrefix(fmt.Sprintf("%s:%s", TokenPrefix, userid))
		uStr := options.Get(c.Request.Context(), vKey)
		if uStr == "" {
			zlog.Warnf("auth token userid=%s not found", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		var auth Authorization
		if err = sonic.UnmarshalString(uStr, &auth); err != nil {
			zlog.Warnf("auth token userid=%s unmarshal err: %v", userid, err)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}
		// 是否允许多设备登录
		if !options.AllowMultipleDevice && auth.Session != zcpt.Md5(token) {
			zlog.Warnf("auth token userid=%s device changed", userid)
			zgin.AbortHttpCode(c, http.StatusUnauthorized, zgin.ErrInvalidToken)
			return
		}

		// 临时存储用户资料
		c.Set(LocalsUserPrefix, zutil.Ptr(auth.Value))
		c.Set(LocalsSessionPrefix, auth.Session)

		// 刷新Token有效期
		c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
		options.Set(c.Request.Context(), vKey, uStr, options.Age)
		c.Next()
	}
}

func Login(c *gin.Context, user Userinfo) zgin.RespBean {
	vKey := options.WithPrefix(fmt.Sprintf("%s:%s", TokenPrefix, user.Userid()))
	// 是否允许多设备登录
	if !options.AllowMultipleDevice {
		options.Delete(c.Request.Context(), vKey)
	}
	// 生成登录态
	tk := fmt.Sprintf("%s##%s##%s##%s##%d", zid.NextIdShort(), zcpt.Md5(c.Request.UserAgent()), c.ClientIP(), user.Userid(), time.Now().Unix())
	d, _ := zcpt.AesEncryptCBC([]byte(tk), []byte(AESKey))
	token := base64.StdEncoding.EncodeToString(d)
	c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
	userStr, _ := sonic.MarshalString(&Authorization{Session: zcpt.Md5(token), Value: user})
	options.Set(c.Request.Context(), vKey, userStr, options.Age)
	return zgin.NewRespWithData(gin.H{
		"token":  token,
		"expire": time.Now().Add(options.Age).Format(time.RFC3339),
	})
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
	uStr, _ := sonic.MarshalString(&Authorization{Session: session.(string), Value: user})
	vKey := options.WithPrefix(fmt.Sprintf("%s:%s", TokenPrefix, user.Userid()))
	options.Set(c.Request.Context(), vKey, uStr, options.Age)
}

func Auth(c *gin.Context) (Userinfo, error) {
	if u, ok := c.Get(LocalsUserPrefix); ok {
		return u.(Userinfo), nil
	}
	return nil, fmt.Errorf("auth not found")
}

func Token(c *gin.Context) string {
	auth, _ := c.Cookie("auth")
	return strings.TrimSpace(zutil.FirstTruth(
		c.GetHeader("Authorization"),
		c.Query("auth"),
		auth,
	))
}
