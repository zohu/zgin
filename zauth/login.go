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
		zch.R().Set(c.Request.Context(), zch.PrefixAuthPreID.Key(id), "waiting", expire)
		return zgin.MessageSuccess.Resp(c).WithData(&Tokens{
			ID:       id,
			Redirect: resp.Redirect,
			Qrcode:   resp.Qrcode,
			Expire:   int64(expire.Seconds()),
		})
	}
	return zgin.MessageLoginUnsupportedMode.Resp(c)
}
func postLogin(c *gin.Context, h *ParamLoginPost) *zgin.RespBean {
	if zch.R().Get(c.Request.Context(), zch.PrefixAuthPreID.Key(h.ID)).Val() == "" {
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
		zch.R().Set(c.Request.Context(), zch.PrefixAuthPreID.Key(h.ID), "done", time.Minute*5)
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
	vKey := zch.PrefixAuthToken.Key(user.Userid())
	// 是否允许多设备登录
	if !options.AllowMultipleDevice {
		zch.R().Del(c.Request.Context(), vKey)
	}
	// 生成登录态
	tk := fmt.Sprintf("%s##%s##%s##%s##%d", zid.NextIdShort(), zcpt.Md5(c.Request.UserAgent()), c.ClientIP(), user.Userid(), time.Now().Unix())
	d, _ := zcpt.AesEncryptCBC([]byte(tk), []byte(AESKey))
	token := base64.StdEncoding.EncodeToString(d)
	c.SetCookie("auth", token, int(options.Age.Seconds()), "", "", false, false)
	userStr, _ := sonic.MarshalString(&Authorization[Userinfo]{Session: zcpt.Md5(token), Value: user})
	zch.R().Set(c.Request.Context(), vKey, userStr, options.Age)
	return zgin.MessageSuccess.Resp(c).WithData(&Tokens{
		Token:  token,
		Expire: int64(options.Age.Seconds()),
	})
}
