package zauth

import (
	"encoding/base64"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"github.com/zohu/zgin/zch"
	"github.com/zohu/zgin/zcpt"
	"github.com/zohu/zgin/zid"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zmap"
	"github.com/zohu/zgin/zutil"
	"time"
)

const (
	PrefixPreID zch.Prefix = "auth:pre"
	PrefixToken zch.Prefix = "auth:user"
)

var logins = zmap.NewStringer[LoginMode, LoginEntity]()

func LoginMethodAdd(mode LoginMode, login LoginEntity) {
	zlog.Infof("add login method: %s", mode)
	logins.Set(mode, login)
}

func LoginRouteRegister(r *gin.RouterGroup) {
	r.POST("/login", zgin.Bind(preLogin))
	r.POST("/token", zgin.Bind(postLogin))
}

func preLogin(c *gin.Context, h *ParamLoginPre) *zgin.RespBean {
	id := zid.NextIdHex()
	if entity, ok := logins.Get(h.Mode); ok {
		resp, err := entity.PreLogin(c, id, h)
		if err != nil {
			return zgin.MessageLoginFailed.Resp(c).AddMessage(err.Error())
		}
		if resp.User != nil && resp.User.Userid() != "" {
			return activeToken(c, resp.User)
		}
		expire := zutil.When(resp.PreExpire > 0, resp.PreExpire, time.Minute*5)
		options.Set(c.Request.Context(), PrefixPreID.Key(id), "waiting", expire)
		return zgin.MessageSuccess.Resp(c).WithData(&Tokens{
			ID:       id,
			Redirect: resp.Redirect,
			Qrcode:   resp.Qrcode,
			Expire:   time.Now().Add(expire).Format(time.RFC3339),
		})
	}
	return zgin.MessageLoginUnsupportedMode.Resp(c)
}
func postLogin(c *gin.Context, h *ParamLoginPost) *zgin.RespBean {
	if options.Get(c.Request.Context(), PrefixPreID.Key(h.ID)) == "" {
		return zgin.MessageLoginTimeout.Resp(c)
	}
	if entity, ok := logins.Get(h.Mode); ok {
		resp, err := entity.PostLogin(c, h.Mode, h.ID)
		if err != nil {
			return zgin.MessageLoginFailed.Resp(c).AddMessage(err.Error())
		}
		if !resp.IsDone {
			return zgin.MessageSuccess.Resp(c).WithData("waiting")
		}
		options.Set(c.Request.Context(), PrefixPreID.Key(h.ID), "done", time.Minute*5)
		if resp.User != nil && resp.User.Userid() != "" {
			return activeToken(c, resp.User)
		}
		return zgin.MessageLoginFailed.Resp(c)
	}
	return zgin.MessageLoginUnsupportedMode.Resp(c)
}
func activeToken(c *gin.Context, user Userinfo) *zgin.RespBean {
	if vali := user.Validate(); vali != zgin.MessageSuccess {
		return vali.Resp(c)
	}
	vKey := PrefixToken.Key(user.Userid())
	// 是否允许多设备登录
	if !options.AllowMultipleDevice {
		options.Delete(c.Request.Context(), vKey)
	}
	// 生成登录态
	tk := fmt.Sprintf("%s##%s##%s##%s##%d", zid.NextIdShort(), zcpt.Md5(c.Request.UserAgent()), c.ClientIP(), user.Userid(), time.Now().Unix())
	d, _ := zcpt.AesEncryptCBC([]byte(tk), []byte(AESKey))
	token := base64.StdEncoding.EncodeToString(d)
	c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
	userStr, _ := sonic.MarshalString(&Authorization[Userinfo]{Session: zcpt.Md5(token), Value: user})
	options.Set(c.Request.Context(), vKey, userStr, options.Age)
	return zgin.MessageSuccess.Resp(c).WithData(&Tokens{
		Token:  token,
		Expire: time.Now().Add(options.Age).Format(time.RFC3339),
	})
}
