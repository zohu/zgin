package zgin

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
)

var (
	ErrParameter      = NewFlag(400, "参数错误")
	ErrInvalidToken   = NewFlag(401, "未登录")
	ErrInvalidSession = NewFlag(401, "已在其他地方登录，请确认账号密码是否泄露")
	ErrNil            = NewFlag(500, "未知错误，联系管理员")
	ErrNotImplemented = NewFlag(501, "暂不支持")
)

func NewFlag(code int, message string) RespBean {
	return NewResp(code, message, nil, nil)
}
func (r RespBean) WithValidateErrs(h interface{}, errs error) RespBean {
	var ves validator.ValidationErrors
	if errors.As(errs, &ves) {
		r.Notes = translateErrors(h, ves)
	} else {
		r.Message = errs.Error()
	}
	return r
}
func (r RespBean) WithMessage(msg string) RespBean {
	r.Message = fmt.Sprintf("%s, %s", r.Message, msg)
	return r
}
