package zgin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"strings"
)

type Pages struct {
	Page int `json:"page" xml:"page" note:"页码"`
	Size int `json:"size" xml:"size" note:"每页数量"`
}

func (p *Pages) PageSizes() (int, int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 50
	}
	if p.Size > 1000 {
		p.Size = 1000
	}
	return p.Page, zutil.FirstTruth(p.Size, 50)
}
func (p *Pages) ScopePage(db *gorm.DB) *gorm.DB {
	page, size := p.PageSizes()
	return db.Offset((page - 1) * size).Limit(size)
}

type MessageID string
type Empty struct{}
type RespBean struct {
	Code    int               `json:"code" xml:"code"`
	Data    any               `json:"data,omitempty" xml:"data"`
	Message string            `json:"message,omitempty" xml:"message"`
	Notes   map[string]string `json:"notes,omitempty" xml:"notes"`
}
type RespListBean[T any] struct {
	Page  int `json:"page" xml:"page"`
	Size  int `json:"size" xml:"size"`
	Total int `json:"total" xml:"total"`
	List  []T `json:"list" xml:"list"`
}

func (m MessageID) Resp(c *gin.Context, kv ...map[string]string) *RespBean {
	resp := &RespBean{
		Code:    1,
		Data:    Empty{},
		Message: string(m),
		Notes:   make(map[string]string),
	}
	if m == MessageSuccess {
		resp.Message = "ok"
	} else {
		arr := strings.Split(string(m), ":")
		if len(arr) == 2 {
			status, _ := strconv.Atoi(arr[0])
			resp.Code = status
			resp.Message = SafeLocalize(c, arr[1], kv...)
		}
	}
	return resp
}

func (r *RespBean) WithValidateErrs(c *gin.Context, h interface{}, errs error) *RespBean {
	var ves validator.ValidationErrors
	if errors.As(errs, &ves) {
		r.Notes = translateErrors(c, h, ves)
	} else {
		r.Message = errs.Error()
	}
	return r
}

func (r *RespBean) AddMessage(msg string) *RespBean {
	r.Message = fmt.Sprintf("%s, %s", r.Message, msg)
	return r
}
func (r *RespBean) WithData(data any) *RespBean {
	r.Data = data
	return r
}

func SafeLocalize(c *gin.Context, ID string, kv ...map[string]string) string {
	if l, ok := c.Get("localizer"); ok {
		data := map[string]string{}
		if len(kv) > 0 {
			data = kv[0]
		}
		message, err := l.(*i18n.Localizer).Localize(&i18n.LocalizeConfig{
			MessageID:    ID,
			TemplateData: data,
		})
		if err == nil {
			return message
		}
		zlog.Warnf("翻译错误: %v", err)
	}
	return ID
}

func translateErrors(c *gin.Context, h any, errs validator.ValidationErrors) map[string]string {
	ets := make(map[string]string)
	elem := reflect.TypeOf(h)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	for _, err := range errs {
		field, _ := elem.FieldByName(err.StructField())
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		if msg := field.Tag.Get("message"); msg != "" {
			ets[key] = SafeLocalize(c, msg, nil)
		} else {
			ets[key] = err.Error()
		}
	}
	return ets
}
