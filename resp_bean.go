package zgin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/z18n"
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

type Empty struct{}
type RespBean struct {
	Code    int               `json:"code" xml:"code"`
	Data    any               `json:"data,omitempty" xml:"data"`
	Message string            `json:"message,omitempty" xml:"message"`
	Notes   map[string]string `json:"notes,omitempty" xml:"notes"`
}
type RespListBean[T any] struct {
	Page  int   `json:"page" xml:"page"`
	Size  int   `json:"size" xml:"size"`
	Total int64 `json:"total" xml:"total"`
	List  []T   `json:"list" xml:"list"`
}
type Option[V any, E any] struct {
	Label    string         `json:"label"`
	Value    V              `json:"value"`
	Extra    E              `json:"extra"`
	Children []Option[V, E] `json:"children"`
}

type MessageID string

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
			resp.Message = z18n.Localize(c, arr[1], kv...)
		}
	}
	return resp
}

func (r *RespBean) WithValidateErrs(c *gin.Context, h interface{}, errs error) *RespBean {
	var ves validator.ValidationErrors
	if errors.As(errs, &ves) {
		r.Notes = translateErrors(c, h, ves)
	} else {
		zlog.Warnf("params err: %v", errs)
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
			ets[key] = z18n.Localize(c, msg)
		} else {
			ets[key] = err.Error()
		}
	}
	return ets
}
