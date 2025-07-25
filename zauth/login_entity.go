package zauth

import (
	"github.com/gin-gonic/gin"
	"github.com/zohu/zgin"
	"time"
)

type LoginMode string

func (m LoginMode) String() string {
	return string(m)
}

type Userinfo interface {
	Userid() string
	UserName() string
	Validate() zgin.MessageID
}

type Authorization[T Userinfo] struct {
	Session string `json:"session"`
	Value   T      `json:"value"`
}
type Tokens struct {
	ID       string `json:"id,omitempty"`
	Redirect string `json:"redirect,omitempty"`
	Qrcode   string `json:"qrcode,omitempty"`
	Token    string `json:"token,omitempty"`
	Expire   int64  `json:"expire,omitempty"`
}
type ParamLoginPre struct {
	Mode    LoginMode `json:"mode" binding:"required" message:"Login.Mode"`
	Account string    `json:"account"`
	Code    string    `json:"code"`
}
type ParamLoginPost struct {
	Mode LoginMode `json:"mode" binding:"required" message:"Login.Mode"`
	ID   string    `json:"id" binding:"required"`
}
type RespLogin struct {
	Tokens
	PreExpire time.Duration // 预登录过期时间
	IsDone    bool          // 登录逻辑是否走完
	User      Userinfo      // 用户信息
}

type LoginEntity interface {
	// PreLogin
	// @Description: 预登陆，如果可以一次性登录则返回用户信息，否则返回预登陆信息Tokens
	// @param c
	// @param h
	// @return RespLogin
	// @return error
	PreLogin(c *gin.Context, ID string, h *ParamLoginPre) (*RespLogin, error)
	// PostLogin
	// @Description: 异步登录认证，必须返回IsDone，否则会超时失败
	// @param c
	// @param h
	// @return RespLogin
	// @return error
	PostLogin(c *gin.Context, mode LoginMode, ID string) (*RespLogin, error)
}
